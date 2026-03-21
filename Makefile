BINARY = dotpack
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS = -ldflags "-s -w -X main.Version=$(VERSION)"

.PHONY: build build-go build-darwin push status install clean test

build-go:
	go build $(LDFLAGS) -o $(BINARY) .

build: build-go
	./$(BINARY) build

build-darwin: build-go
	./$(BINARY) build --os darwin

push: build-go
	./$(BINARY) push $(HOST)

status: build-go
	./$(BINARY) status $(HOST)

install: build-go
	./$(BINARY) install

clean:
	rm -f $(BINARY) dotpack-*.tar.gz
	-docker rmi dotpack 2>/dev/null

test: build-go
	./$(BINARY) version
	./$(BINARY) versions
