
# skipctl

A simple client (and [server](./server.md)) to perform various network troubleshooting.

![skipctl](assets/skipctl-logo.png)

- [Installation](#installation)
- [Usage](#usage)
    - [Test](#test)
        - [Ping](#ping)
        - [Port probe](#port-probe)
    - [Manifests](#manifests)
        - [Render manifests](#render-manifests)
        - [Validate manifests](#validate-k8s-manifests)
- [Analytics & Privacy](#analytics--privacy)
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

The `manifests` command group contains commands for working with Skiperator manifests. All commands require a path pointing to a directory or file containing the manifest(s).
If a directory is specified, all commands in this group will recursively search the specified path for files with supported file formats. If no specific path is specified, the current working directory will be used.

**Supported file formats are:**
- `.jsonnet`
- `.libsonnet` (only for `manifests format`)
- `.yaml`
- `.yml`
- `kustomization.yaml`
- `kustomization.yml`


#### Render manifests
>[!note]
> When rendering kustomize, the renderer will ignore other `.jsonnet` and `.yaml` files in the same directory and subdirectories.

Compiles and renders a Skiperator `.jsonnet` or `.yaml` manifest in the specified directory and alerts if any errors are found.
```shell
skipctl manifests render <pathname>
```
#### Format manifests

Formats a `.jsonnet`, `.libsonnet` or `.yaml` manifest in the specified directory and alerts if any errors are found.
```shell
skipctl manifests format <pathname>
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
skipctl manifests format .

# Format a single file
skipctl manifests format ./app/manifest.yaml

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
--ref <git-ref>        Git ref or directory to diff against (default HEAD)
--diff-format <fmt>    pretty | patch | json (default pretty)
--verbosity <level>    full | chunk | minimal (auto-set if omitted)
--chunk-size <n>       Context lines for chunk (default 3)
--path, -p <path>      Files / directory to scan
--kustomize            Enable diffing kustomize directories (requires --ref to be a directory)
--sort-output          Sort rendered outputs before diffing (useful on refactors that change output order only)
```

Defaults (when --verbosity not provided): pretty->full, patch->chunk, json->full.

Examples:
```shell
skipctl manifests diff --path .
skipctl manifests diff --path . --diff-format json --verbosity minimal | jq '.diffs[] | select(.type!="Equals")'
skipctl manifest diff --path env/cluster --verbosity chunk --diff-format patch --kustomize --ref env-copy/cluster 
```

Exit codes: 0 success / 1 error.



#### Validate K8s manifests

Validates manifest files (in either `.jsonnet` or `.yaml` format) against relevant CRDs for Kartverket, 
including CRDs defined in [Skiperator](https://github.com/kartverket/skiperator), [Ztoperator](https://github.com/kartverket/ztoperator) 
and [Accesserator](https://github.com/kartverket/accesserator). 
The complete list of supported CRDs can be viewed [here](./pkg/crd/schemas).

The API references for Skiperator and Ztoperator can be found in the SKIP docs, 
respectively [here](https://skip.kartverket.no/docs/applikasjon-utrulling/skiperator/api-docs) and [here](https://skip.kartverket.no/docs/applikasjon-utrulling/tilgangsstyring/ztoperator/api-docs).

Returns status code `1` if there are failures or `0` for successful validation

```shell
skipctl manifests validate <pathname>
```

## Analytics & Privacy

`skipctl` collects anonymous usage analytics by **default** to help us understand how the tool is being used and improve the user experience.

### What We Collect

We collect the following **non-personal** information:
- **Command usage**: Which commands and subcommands are executed (e.g., `test ping`, `manifests render`)
- **Command arguments**: Arguments passed to commands (e.g., flags used)
- **Error information**: Whether a command succeeded or failed, and error messages if applicable. The error message may itself contain information about your system, e.g. file paths.
- **Environment context**:
  - Operating system (e.g., macOS, Linux, Windows)
  - System architecture (e.g., amd64, arm64)
  - Application version and git commit hash of skipctl
  - Environment type (local or CI)
- **Anonymous identifier**: A hashed machine-specific identifier that cannot be traced back to you.

### How We Collect Data

Analytics are collected through [PostHog](https://posthog.com/), a privacy-focused analytics platform, and sent to our self-hosted instance.

### How to Disable Analytics

You can disable analytics collection in two ways:

1. Set the `DO_NOT_TRACK` environment variable to `true`
2. Use the `--no-analytics` flag when running any `skipctl` command

You can suppress analytics enablement messages (whether it enabled or not) by setting the environement variable `SKIPCTL_MUTE_TELEMETRY` to `true`. 
