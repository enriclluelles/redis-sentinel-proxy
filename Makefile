GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
CGO_ENABLED ?= 0

.PHONY: test
test:
	go test -v ./...
	cd test && ./test.sh

.PHONY: build
build:
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags '-s -w -extldflags "-static"' -o bin/redis-sentinel-proxy .