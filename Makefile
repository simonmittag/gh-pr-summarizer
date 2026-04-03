.PHONY: all build test clean

BINARY_NAME=gh-pr-summarizer
MAIN_PACKAGE=./cmd/gh-pr-summarizer

all: build test

build:
	go build -o $(BINARY_NAME) $(MAIN_PACKAGE)

test:
	go test ./...

clean:
	rm -f $(BINARY_NAME)
