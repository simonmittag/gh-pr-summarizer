.PHONY: all build test clean

BINARY_NAME=gh-pr-summarizer
MAIN_PACKAGE=./cmd/gh-pr-summarizer
VERSION ?= $(shell git describe --tags --always --dirty || echo "dev")
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

all: build test

build:
	go build $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PACKAGE)

test:
	go test ./...

clean:
	rm -f $(BINARY_NAME)
