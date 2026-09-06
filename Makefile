# Makefile for md2pdf.
#
# VERSION has a fixed default and can be overridden on the command line:
#
#   make build VERSION=2.4.0

VERSION ?= dev
LDFLAGS := -s -w -X github.com/gstuebner/md2pdf/cmd.version=$(VERSION)

DIST := dist
BINARY := md2pdf

.PHONY: build build-all test fmt vet clean

build:
	mkdir -p $(DIST)
	go build -ldflags "$(LDFLAGS)" -o $(DIST)/$(BINARY) .

build-all:
	mkdir -p $(DIST)
	GOOS=linux   GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(DIST)/$(BINARY)_linux_amd64 .
	GOOS=linux   GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(DIST)/$(BINARY)_linux_arm64 .
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(DIST)/$(BINARY)_windows_amd64.exe .

test:
	go test ./...

fmt:
	gofmt -l .

vet:
	go vet ./...

clean:
	rm -rf $(DIST)
