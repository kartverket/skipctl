# Calculate version
version ?= $(shell git describe --always --tags --dirty)

debug:
	go build -race -ldflags="-X 'main.GitCommitHash=$(version)'"

build:
	go build -trimpath -ldflags="-s -w -X 'main.GitCommitHash=$(version)'"

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
