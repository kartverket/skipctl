//go:build generate

package crd

// Create an alias for the tool command to generate OpenAPI schemas from CRD files.
//go:generate -command crdjson go run ../../hack/crd2openapi

// Skiperator CRDs
//go:generate crdjson -outdir=./schemas -url=https://raw.githubusercontent.com/kartverket/skiperator/refs/heads/main/config/crd/skiperator.kartverket.no_applications.yaml
//go:generate crdjson -outdir=./schemas -url=https://raw.githubusercontent.com/kartverket/skiperator/refs/heads/main/config/crd/skiperator.kartverket.no_skipjobs.yaml
//go:generate crdjson -outdir=./schemas -url=https://raw.githubusercontent.com/kartverket/skiperator/refs/heads/main/config/crd/skiperator.kartverket.no_routings.yaml

// Skiperator dependencies down below
//go:generate crdjson -outdir=./schemas -url=https://raw.githubusercontent.com/nais/liberator/refs/heads/main/config/crd/bases/nais.io_azureadapplications.yaml
//go:generate crdjson -outdir=./schemas -url=https://raw.githubusercontent.com/nais/liberator/refs/heads/main/config/crd/bases/nais.io_idportenclients.yaml
//go:generate crdjson -outdir=./schemas -url=https://raw.githubusercontent.com/nais/liberator/refs/heads/main/config/crd/bases/nais.io_maskinportenclients.yaml

//go:generate crdjson -outdir=./schemas -url=https://github.com/prometheus-operator/prometheus-operator/releases/latest/download/stripped-down-crds.yaml

//go:generate crdjson -outdir=./schemas -url=https://github.com/prometheus-operator/prometheus-operator/releases/latest/download/stripped-down-crds.yaml

//go:generate crdjson -outdir=./schemas -one-per-kind -url=https://raw.githubusercontent.com/istio/api/refs/tags/1.27.1/kubernetes/customresourcedefinitions.gen.yaml

//go:generate crdjson -outdir=./schemas -url=https://github.com/cert-manager/cert-manager/releases/latest/download/cert-manager.crds.yaml

//go:generate crdjson -outdir=./schemas -one-per-kind -url=https://raw.githubusercontent.com/external-secrets/external-secrets/refs/heads/main/deploy/crds/bundle.yaml

//go:generate crdjson -outdir=./schemas -url=https://raw.githubusercontent.com/cloudnative-pg/cloudnative-pg/release-1.27/releases/cnpg-1.27.0.yaml
