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

//go:generate crdjson -outdir=./schemas -one-per-kind -url=https://raw.githubusercontent.com/istio/api/refs/tags/1.28.0/kubernetes/customresourcedefinitions.gen.yaml

//go:generate crdjson -outdir=./schemas -url=https://github.com/cert-manager/cert-manager/releases/download/v1.19.1/cert-manager.crds.yaml

//go:generate crdjson -outdir=./schemas -one-per-kind -url=https://github.com/external-secrets/external-secrets/releases/download/v1.0.0/external-secrets.yaml

//go:generate crdjson -outdir=./schemas -url=https://github.com/cloudnative-pg/cloudnative-pg/releases/download/v1.27.1/cnpg-1.27.1.yaml

// Ztoperator CRD
//go:generate crdjson -outdir=./schemas -url=https://raw.githubusercontent.com/kartverket/ztoperator/refs/heads/main/config/crd/bases/ztoperator.kartverket.no_authpolicies.yaml

// Accesserator CRD
//go:generate crdjson -outdir=./schemas -url=https://raw.githubusercontent.com/kartverket/accesserator/refs/heads/main/config/crd/bases/accesserator.kartverket.no_securityconfigs.yaml

// Accesserator dependencies down below
//go:generate crdjson -outdir=./schemas -url=https://raw.githubusercontent.com/nais/liberator/34a9e8ad9d4bf1b6f8e292e05ccb71db1147fbcb/config/crd/bases/nais.io_jwkers.yaml
