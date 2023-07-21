GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
CGO_ENABLED ?= 0

.PHONY: tests-intergration
tests-intergration:
	cd test && ./test.sh

.PHONY: tests-unit
tests-unit:
	CGO_ENABLED=1 go test -v -race -cover ./...

.PHONY: build
build:
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags '-s -w -extldflags "-static"' -o bin/redis-sentinel-proxy .