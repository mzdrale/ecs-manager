BINARY := ecs-manager
VERSION ?= $(shell git describe --abbrev=0 --tags)

.PHONY: build
build:
	@make build:linux
	@make build:darwin

.PHONY: build\:darwin
build\:darwin:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o ${BINARY}-${VERSION}-darwin-amd64 -ldflags="-s -w -X main.binName=${BINARY} -X main.version=${VERSION}"

.PHONY: build\:linux
build\:linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${BINARY}-${VERSION}-linux-amd64 -ldflags="-s -w -X main.binName=${BINARY} -X main.version=${VERSION}"

.PHONY: run
run:
	@go run .

.PHONY: help
help:
	@echo "build   		- Compile go code and provide binary for macOS and Linux"
	@echo "build:linux 	- Compile go code and provide binary for Linux"
	@echo "build:darwin - Compile go code and provide binary for macOS (Darwin)"
	@echo "run     		- Compile and run go code"
