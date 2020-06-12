BINARY := ecs-manager
VERSION := $(shell cat VERSION)
REVISION := $(shell git rev-parse --short HEAD)
CONFIG_DIR := ~/.config/ecs-manager
CONFIG := config.toml

.PHONY: build
build:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o ${BINARY} -ldflags="-s -w -X main.binName=${BINARY} -X main.version=${VERSION}-${REVISION}"

.PHONY: install
install:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go install -ldflags="-s -w -X main.binName=${BINARY} -X main.version=${VERSION}-${REVISION}"
	@test -d ${CONFIG_DIR}/${CONFIG} || mkdir -p ${CONFIG_DIR} && cp ${CONFIG} ${CONFIG_DIR}/${CONFIG}

.PHONY: run
run:
	@go run .

.PHONY: help
help:
	@echo "build   - Compile go code and provide binary for macOS"
	@echo "install - Compile and install binary in ${GOPATH}/bin"
	@echo "run     - Compile and run go code"