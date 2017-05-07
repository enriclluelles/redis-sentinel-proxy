IMAGE_NAME := anubhavmishra/redis-sentinel-proxy
.PHONY: test

.DEFAULT_GOAL := help
help: ## List targets & descriptions
	@cat Makefile* | grep -E '^[a-zA-Z_-]+:.*?## .*$$' | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

deps:
	go get .

run-docker: ## Run dockerized service directly
	docker run $(IMAGE_NAME):latest

push: ## docker push image to registry
	docker push $(IMAGE_NAME):latest

build: ## Build the project
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build .
	docker build -t $(IMAGE_NAME):latest .

run: ## Build and run the project
	go build . && ./redis-sentinel-proxy

clean:
	-rm -rf build
