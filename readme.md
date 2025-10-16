
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
#### Format manifests

Formats a `.jsonnet` or `.yaml` manifest in the specified directory and alerts if any errors are found.
```shell
skipctl manifests format --path <pathname>
```

##### Format quick guide

Purpose: Rewrite manifests in-place to a canonical style (JSONNet / YAML). Recurses a directory or reads from stdin.

Inputs:
```
--path, -p <path>   Directory or file (defaults to CWD)
(stdin)             Use '-' as sole argument to read from standard input
```

Behavior:
- Only files with supported suffixes are touched (`.jsonnet`, `.yaml`, `.yml`).
- On error (parse / write) returns exit code 1 after logging.

Examples:
```shell
# Format everything under current directory
skipctl manifests format --path .

# Format a single file
skipctl manifests format --path ./app/manifest.yaml

# Format from stdin (outputs the formatted content to stdout)
cat manifest.yaml | skipctl manifests format -
```

Exit codes: 0 success / 1 error.
#### Diff manifests

Compares a Skiperator manifest against the currently deployed version in a Kubernetes cluster and shows the differences.

```shell
skipctl manifests diff --path <pathname> [flags]
```


##### Diff quick guide

Flags:
```
--ref <git-ref>        Git ref to diff against (default HEAD)
--diff-format <fmt>    pretty | patch | json (default pretty)
--verbosity <level>    full | chunk | minimal (auto-set if omitted)
--chunk-size <n>       Context lines for chunk (default 3)
--path, -p <path>      Files / directory to scan
```

Defaults (when --verbosity not provided): pretty->full, patch->chunk, json->full.

Examples:
```shell
skipctl manifests diff --path .
skipctl manifests diff --path . --diff-format json --verbosity minimal | jq '.diffs[] | select(.type!="Equals")'
```

Exit codes: 0 success / 1 error.



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
