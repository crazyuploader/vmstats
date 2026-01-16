.PHONY: build run test clean

BINARY_NAME=vmstats

build:
	@mkdir -p bin
	go build -o bin/$(BINARY_NAME) cmd/vmstats/main.go

run: build
	./bin/$(BINARY_NAME)

test:
	go test -v ./...

clean:
	rm -f bin/$(BINARY_NAME)
	rm -f vmstats.log
