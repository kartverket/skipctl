
# skipctl

A simple client (and [server](./server.md)) to perform various network troubleshooting.

- [Installation](#installation)
- [Usage](#usage)
    - [Test](#test)
        - [Ping](#ping)
        - [Port probe](#port-probe)
    - [Manifests](#manifests)
        - [Render manifests](#render-manifests)
        - [Validate manifests](#validate-k8s-manifests)

## Installation

Download the [latest release](https://github.com/kartverket/skipctl/releases) or use the provided Docker image (mainly for running a server).

### Use Homebrew

```shell
brew tap kartverket/taps && \
brew install skipctl
```

## Usage

### Test
The various `test` commands will execute against an API server. Run `skipctl test ping --api-server=something` to get a list of supported API server names.
An API server represents a location that can run tests from their perspective. All communication with API servers is encrypted over TLS.

> :exclamation: Before issuing any `test` commands, be sure to be authenticated first (`gcloud auth application-default login`).


#### Ping

```shell
skipctl test ping --hostname=example.com --api-server=myApiServer
```

#### Port probe

```shell
skipctl test probe --hostname=example.com --port=1521 --api-server=myApiServer
```


### Manifests

The `manifests` command group contains commands for working with Skiperator manifests. All commands require a path flag (`--path` or `-p`) pointing to a directory or file containing the manifest(s).
If a directory is specified, all commands in this group will recursively search the specified path for files with supported file formats. If no specific path is specified, the current working directory will be used.  

**Supported file formats are:**
- `.jsonnet`
- `.yaml`
- `.yml`

#### Render manifests

Compiles and renders a Skiperator `.jsonnet` or `.yaml` manifest in the specified directory and alerts if any errors are found.
```shell
skipctl manifests render --path <pathname>
```

#### Validate K8s manifests

Validates a Skiperator manifest file (in either `.jsonnet` or `.yaml` format) against Skiperator's own custom schema definitions (skiperator.kartverket.no/v1alpha1)

**Supports the following resource types:**
- Application
- Routing
- SKIPJob

The API reference for these types can be found in the [SKIP docs](https://skip.kartverket.no/docs/applikasjon-utrulling/skiperator/api-docs).

Returns status code `1` if there are failures or `0` for successful validation

```shell
skipctl manifests validate --path <pathname>
```
