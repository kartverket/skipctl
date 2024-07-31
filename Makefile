.DEFAULT_GOAL = build

export CGO_ENABLED = 0

# Calculate version
version ?= $(shell git describe --always --tags --dirty)

.PHONY: test
test:
	go test ./...

.PHONY: lint
lint:
	golangci-lint run --fix --timeout 10m

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: proto-lint
proto-lint:
	cd proto/ && go run github.com/bufbuild/buf/cmd/buf lint

.PHONY: generate
generate:
	rm -rf ./pkg/api
	go run github.com/bufbuild/buf/cmd/buf generate proto

.PHONY: debug
debug:
	go build -race -ldflags="-X 'main.GitCommitHash=$(version)'"

.PHONY: build
build: lint test proto-lint generate
	go build -tags osusergo,netgo -trimpath -ldflags="-s -w -X 'main.GitCommitHash=$(version)'"

.PHONY: build-nolint
build-nolint: test proto-lint generate
	go build -tags osusergo,netgo -trimpath -ldflags="-s -w -X 'main.GitCommitHash=$(version)'"
