# Calculate version
version ?= $(shell git describe --always --tags --dirty)

build:
	go build

dev: lint test build

test:
	go test ./...

lint:
	golangci-lint run --fix --timeout 10m

fmt:
	go fmt ./...

proto-lint:
	cd proto/ && go run github.com/bufbuild/buf/cmd/buf lint

generate:
	rm -rf ./pkg/api
	go run github.com/bufbuild/buf/cmd/buf generate proto
