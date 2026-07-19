package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const cacheTTL = 24 * time.Hour

const httpTimeout = 30 * time.Second

const retryBase = 500 * time.Millisecond

const retryCap = 8 * time.Second

const maxRetries = 6

type cache struct {
	client *http.Client
	dir    string
}

func newCache() (*cache, error) {
	dir := ".cache"
	err := os.MkdirAll(dir, 0o755)
	if err != nil {
		return nil, fmt.Errorf("create cache dir: %w", err)
	}

	instance := &cache{
		client: &http.Client{Timeout: httpTimeout},
		dir:    dir,
	}
	return instance, nil
}

func (c *cache) path(url string) string {
	sum := sha256.Sum256([]byte(url))
	name := fmt.Sprintf("%x.json", sum)
	return filepath.Join(c.dir, name)
}

func (c *cache) get(url string, force bool) ([]byte, error) {
	file := c.path(url)

	if !force {
		info, err := os.Stat(file)
		if err == nil && time.Since(info.ModTime()) < cacheTTL {
			body, e := os.ReadFile(file)
			if e == nil {
				return body, nil
			}
		}
	}

	body, err := c.download(url)
	if err != nil {
		stale, readErr := os.ReadFile(file)
		if readErr != nil {
			return nil, err
		}
		fmt.Fprintf(os.Stderr, "f1: %v; serving stale cache\n", err)
		return stale, nil
	}

	err = os.WriteFile(file, body, 0o644)
	if err != nil {
		return nil, fmt.Errorf("write cache %s: %w", url, err)
	}
	return body, nil
}

// download fetches url, retrying on rate-limit (429) and transient 5xx
// responses with exponential backoff that honours any Retry-After header.
func (c *cache) download(url string) ([]byte, error) {
	lastErr := error(nil)

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(backoff(attempt, lastErr))
		}

		body, retry, err := c.attempt(url)
		if err == nil {
			return body, nil
		}

		lastErr = err
		if !retry {
			return nil, err
		}
	}

	return nil, fmt.Errorf("fetch %s: %w", url, lastErr)
}

func (c *cache) attempt(url string) ([]byte, bool, error) {
	response, err := c.client.Get(url)
	if err != nil {
		return nil, true, fmt.Errorf("fetch %s: %w", url, err)
	}
	defer func() {
		e := response.Body.Close()
		if e != nil {
			fmt.Fprintf(os.Stderr, "f1: close body: %v\n", e)
		}
	}()

	if response.StatusCode == http.StatusTooManyRequests || response.StatusCode >= http.StatusInternalServerError {
		wait := retryAfter(response.Header.Get("Retry-After"))
		return nil, true, &retryableError{status: response.StatusCode, url: url, wait: wait}
	}

	if response.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("fetch %s: status %d", url, response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, true, fmt.Errorf("read %s: %w", url, err)
	}
	return body, false, nil
}

type retryableError struct {
	status int
	url    string
	wait   time.Duration
}

func (e *retryableError) Error() string {
	return fmt.Sprintf("fetch %s: status %d", e.url, e.status)
}

func backoff(attempt int, err error) time.Duration {
	target := &retryableError{}
	if errors.As(err, &target) && target.wait > 0 {
		return target.wait
	}

	wait := retryBase * time.Duration(1<<(attempt-1))
	if wait > retryCap {
		return retryCap
	}
	return wait
}

func retryAfter(header string) time.Duration {
	seconds, err := strconv.Atoi(header)
	if err != nil || seconds <= 0 {
		return 0
	}
	return time.Duration(seconds) * time.Second
}
