BINARY := ecs-manager-ng
REVISION := $(shell git rev-parse --short HEAD)

.PHONY: build
build:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o ${BINARY}-darwin-amd64 -ldflags="-s -w -X main.binName=${BINARY} -X main.version=${REVISION}"
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${BINARY}-linux-amd64 -ldflags="-s -w -X main.binName=${BINARY} -X main.version=${REVISION}"

.PHONY: run
run:
	@go run .

.PHONY: help
help:
	@echo "build   - Compile go code and provide binary for macOS and Linux"
	@echo "run     - Compile and run go code"