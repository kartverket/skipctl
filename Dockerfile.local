FROM golang:1.22 AS builder
WORKDIR /build

# Install golangci-lint
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.59.1

COPY go.mod go.sum ./
RUN go mod download
COPY Makefile ./

COPY . .
RUN make

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/skipctl ./

USER 150:150
ENTRYPOINT ["/skipctl"]
