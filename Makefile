.PHONY: all test clean

test:
	go clean -testcache
	go test -v ./... -timeout 1h

GO_SOURCES := $(shell find . -name "*go" -type f -print)
build: $(GO_SOURCES)
	go build -o build/ .

run: $(GO_SOURCES)
	go run .
