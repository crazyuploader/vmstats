.PHONY: build run test clean lint

BINARY_NAME=vmstats
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT=$(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-s -w -X 'github.com/crazyuploader/vmstats/internal/version.Version=$(VERSION)' \
        -X 'github.com/crazyuploader/vmstats/internal/version.GitCommit=$(GIT_COMMIT)' \
        -X 'github.com/crazyuploader/vmstats/internal/version.BuildDate=$(BUILD_DATE)'

build:
	@mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME) ./cmd/vmstats

run: build
	./bin/$(BINARY_NAME)

test:
	go test -v ./...

lint:
	golangci-lint run

clean:
	rm -rf bin/
	rm -f vmstats.log
