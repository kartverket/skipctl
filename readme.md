# skipctl

A simple client (and [server](./server.md)) to perform various network troubleshooting.

## Installation

Download the [latest release](https://github.com/kartverket/skipctl/releases) or use the provided Docker image (mainly for running a server).

### Use Homebrew

```shell
brew tap kartverket/taps && \
brew install skipctl
```

## Usage

The various `test` commands will execute against an API server. Run `skipctl test ping --api-server=something` to get a list of supported API server names.
An API server represents a location that can run tests from their perspective. All communication with API servers is encrypted over TLS.

> :exclamation: Before issuing any commands, be sure to be authenticated first (`gcloud auth application-default login`). 

### Ping

```shell
skipctl test ping --hostname=example.com --api-server=myApiServer
```

### Port probe

```shell
skipctl test probe --hostname=example.com --port=1521 --api-server=myApiServer
```
