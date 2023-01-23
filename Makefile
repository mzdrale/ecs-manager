BINARY := ecs-manager
VERSION ?= $(shell git describe --abbrev=0 --tags)

.PHONY: build
build:
	@make build:linux
	@make build:macos

.PHONY: build\:macos
build\:macos:
	go get -d
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o ${BINARY}-${VERSION}-macos-amd64 -ldflags="-s -w -X main.binName=${BINARY} -X main.version=${VERSION}"
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o ${BINARY}-${VERSION}-macos-arm64 -ldflags="-s -w -X main.binName=${BINARY} -X main.version=${VERSION}"

.PHONY: build\:linux
build\:linux:
	go get -d
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${BINARY}-${VERSION}-linux-amd64 -ldflags="-s -w -X main.binName=${BINARY} -X main.version=${VERSION}"
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o ${BINARY}-${VERSION}-linux-arm64 -ldflags="-s -w -X main.binName=${BINARY} -X main.version=${VERSION}"

.PHONY: run
run:
	@go run .

.PHONY: help
help:
	@echo "build   		- Compile go code and provide binary for macOS and Linux"
	@echo "build:linux 	- Compile go code and provide binary for Linux"
	@echo "build:macos  - Compile go code and provide binary for macOS (amd64 and arm64)"
	@echo "run     		- Compile and run go code"
