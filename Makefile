.PHONY: build force lint run tidy

BINARY := main

build:
	goimports -w .
	go mod tidy
	go build -o $(BINARY) .

force:
	@:

lint:
	golangci-lint run

run: build
	./$(BINARY) $(if $(filter force,$(MAKECMDGOALS)),--force,)
