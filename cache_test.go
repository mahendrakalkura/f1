package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestBackoff(t *testing.T) {
	cases := []struct {
		name    string
		attempt int
		err     error
		want    time.Duration
	}{
		{"first retry", 1, nil, 500 * time.Millisecond},
		{"second retry", 2, nil, time.Second},
		{"grows exponentially", 4, nil, 4 * time.Second},
		{"capped", 10, nil, retryCap},
		{"honours Retry-After", 3, &retryableError{wait: 3 * time.Second}, 3 * time.Second},
		{"zero wait falls back", 3, &retryableError{wait: 0}, 2 * time.Second},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := backoff(c.attempt, c.err); got != c.want {
				t.Errorf("backoff(%d, %v) = %v, want %v", c.attempt, c.err, got, c.want)
			}
		})
	}
}

func TestCacheGetDownloadError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	store := &cache{client: server.Client(), dir: t.TempDir()}
	_, err := store.get(server.URL+"/missing", false)
	if err == nil || !strings.Contains(err.Error(), "status 404") {
		t.Errorf("got %v, want a status 404 error", err)
	}
}

func TestCacheGetExpired(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, "fresh")
	}))
	defer server.Close()

	store := &cache{client: server.Client(), dir: t.TempDir()}
	file := store.path(server.URL)
	if err := os.WriteFile(file, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	stale := time.Now().Add(-cacheTTL - time.Hour)
	if err := os.Chtimes(file, stale, stale); err != nil {
		t.Fatal(err)
	}

	body, err := store.get(server.URL, false)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "fresh" {
		t.Errorf("got %q, want %q", body, "fresh")
	}
}

func TestCacheGetFetchesAndCaches(t *testing.T) {
	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		_, _ = fmt.Fprint(w, "body")
	}))
	defer server.Close()

	store := &cache{client: server.Client(), dir: t.TempDir()}
	for range 2 {
		body, err := store.get(server.URL, false)
		if err != nil {
			t.Fatal(err)
		}
		if string(body) != "body" {
			t.Errorf("got %q, want %q", body, "body")
		}
	}
	if hits != 1 {
		t.Errorf("server hit %d times, want 1", hits)
	}
}

func TestCacheGetForce(t *testing.T) {
	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		_, _ = fmt.Fprint(w, "body")
	}))
	defer server.Close()

	store := &cache{client: server.Client(), dir: t.TempDir()}
	for range 2 {
		if _, err := store.get(server.URL, true); err != nil {
			t.Fatal(err)
		}
	}
	if hits != 2 {
		t.Errorf("server hit %d times, want 2", hits)
	}
}

func TestCacheGetStaleFallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	store := &cache{client: server.Client(), dir: t.TempDir()}
	file := store.path(server.URL)
	if err := os.WriteFile(file, []byte("cached"), 0o644); err != nil {
		t.Fatal(err)
	}
	stale := time.Now().Add(-cacheTTL - time.Hour)
	if err := os.Chtimes(file, stale, stale); err != nil {
		t.Fatal(err)
	}

	body, err := store.get(server.URL, false)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "cached" {
		t.Errorf("got %q, want stale cache %q", body, "cached")
	}
}

func TestCachePath(t *testing.T) {
	store := &cache{dir: t.TempDir()}
	if store.path("a") == store.path("b") {
		t.Error("different URLs must not share a cache file")
	}
}

func TestCachePrune(t *testing.T) {
	dir := t.TempDir()
	store := &cache{dir: dir}

	fresh := dir + "/fresh.json"
	if err := os.WriteFile(fresh, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	old := dir + "/old.json"
	if err := os.WriteFile(old, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	stale := time.Now().Add(-cacheMaxAge - time.Hour)
	if err := os.Chtimes(old, stale, stale); err != nil {
		t.Fatal(err)
	}

	subdir := dir + "/nested"
	if err := os.Mkdir(subdir, 0o755); err != nil {
		t.Fatal(err)
	}

	store.prune()

	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Error("old file should be pruned")
	}
	if _, err := os.Stat(fresh); err != nil {
		t.Error("fresh file should be kept")
	}
	if _, err := os.Stat(subdir); err != nil {
		t.Error("directories should be left alone")
	}
}

func TestDownloadRetries(t *testing.T) {
	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = fmt.Fprint(w, "recovered")
	}))
	defer server.Close()

	store := &cache{client: server.Client(), dir: t.TempDir()}
	body, err := store.get(server.URL, false)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "recovered" {
		t.Errorf("got %q, want %q", body, "recovered")
	}
	if hits != 2 {
		t.Errorf("server hit %d times, want 2", hits)
	}
}

func TestRetryAfter(t *testing.T) {
	cases := []struct {
		in   string
		want time.Duration
	}{
		{"5", 5 * time.Second},
		{"0", 0},
		{"-3", 0},
		{"abc", 0},
		{"", 0},
	}
	for _, c := range cases {
		if got := retryAfter(c.in); got != c.want {
			t.Errorf("retryAfter(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestRetryableErrorMessage(t *testing.T) {
	err := &retryableError{status: 503, url: "https://example.com"}
	if err.Error() != "fetch https://example.com: status 503" {
		t.Errorf("unexpected message %q", err.Error())
	}
}
