# Authentication in skipctl

This document explains how authentication works in skipctl using Google Cloud service accounts.

## Overview

skipctl uses **Google Cloud ID tokens** from **service accounts only** for authentication. This provides:
- Consistent authentication mechanism for dev, staging, and production
- Better security with key rotation and audit trails
- Native support for GKE Workload Identity
- No dependency on individual user credentials

## How It Works

### Client Side

The client uses service account credentials to obtain ID tokens:

```go
// Uses idtoken.NewTokenSource which discovers credentials from:
// 1. GOOGLE_APPLICATION_CREDENTIALS environment variable (service account key file)
// 2. GCE/GKE metadata server (attached service account or Workload Identity)
```

The client automatically:
1. Gets an ID token from the service account
2. Sends it as a Bearer token in the `authorization` metadata header
3. Establishes a TLS connection to the server

### Server Side

The server validates ID tokens and authenticates service accounts:

#### Service Account Authentication
- **Requirements**: 
  - Valid Google ID token
  - `email` claim present and ends with `.gserviceaccount.com`
- **Use case**: All environments (development, staging, production)
- **Example**: `skipctl@kv-spire-devex-ksde.iam.gserviceaccount.com`

## Setup Instructions

### Development (Local Machine)

1. **Create a service account key**:
   ```bash
   gcloud iam service-accounts keys create ~/.config/gcloud/skipctl-key.json \
     --iam-account=skipctl@kv-spire-devex-ksde.iam.gserviceaccount.com
   ```

2. **Set environment variable**:
   ```bash
   export GOOGLE_APPLICATION_CREDENTIALS="$HOME/.config/gcloud/skipctl-key.json"
   ```

3. **Start the server**:
   ```bash
   ./skipctl serve --gcp-project-id=kv-spire-devex-ksde \
     --tls-cert=$HOME/.config/skipctl/server-cert.pem \
     --tls-key=$HOME/.config/skipctl/server-key.pem
   ```

4. **Run client commands**:
   ```bash
   ./skipctl refactor app.jsonnet
   ```

### Production (Using Service Account Key File)

1. **Create a service account key**:
   ```bash
   gcloud iam service-accounts keys create /path/to/skipctl-key.json \
     --iam-account=skipctl@kv-spire-devex-ksde.iam.gserviceaccount.com
   ```

2. **Set environment variable**:
   ```bash
   export GOOGLE_APPLICATION_CREDENTIALS="/path/to/skipctl-key.json"
   ```

3. **Start the server** (with production TLS certificates):
   ```bash
   ./skipctl serve --gcp-project-id=kv-spire-devex-ksde \
     --tls-cert=/etc/skipctl/server-cert.pem \
     --tls-key=/etc/skipctl/server-key.pem
   ```

4. **Run client commands**:
   ```bash
   ./skipctl refactor app.jsonnet --server=skipctl.example.com:3514
   ```

### Production (Using Workload Identity in GKE)

When running in GKE with Workload Identity configured:

1. **Configure Workload Identity** for your pods (no key file needed)

2. **No environment variables needed** - the service account is attached automatically

3. **Start the server**:
   ```bash
   ./skipctl serve --gcp-project-id=kv-spire-devex-ksde \
     --tls-cert=/etc/skipctl/server-cert.pem \
     --tls-key=/etc/skipctl/server-key.pem
   ```

4. **Run client commands**:
   ```bash
   ./skipctl refactor app.jsonnet --server=skipctl-internal.svc.cluster.local:3514
   ```

## Security Features

### TLS Encryption

- **Localhost**: Uses TLS with `InsecureSkipVerify` for self-signed certificates
- **Remote servers**: Uses TLS with proper certificate validation via system root CAs
- **Server**: Supports TLS 1.2+ with configurable certificates

### Token Validation

The server performs comprehensive validation:

1. **Token format**: Validates Bearer token structure
2. **Signature**: Verifies token signature using Google's public keys
3. **Expiry**: Checks token hasn't expired
4. **Claims**: Validates `email` claim with `.gserviceaccount.com` suffix

### Audit Logging

All authentication attempts are logged with:
- Service account email
- IP address (from gRPC peer info)
- Success/failure status
- Rejection reasons for failed attempts

## Troubleshooting

### "invalid token" Error

**Symptoms**: Client gets `rpc error: code = Unauthenticated desc = invalid token`

**Causes and Solutions**:

1. **Not authenticated / No credentials set**:
   ```bash
   # Solution: Set GOOGLE_APPLICATION_CREDENTIALS
   export GOOGLE_APPLICATION_CREDENTIALS="$HOME/.config/gcloud/skipctl-key.json"
   ```

2. **Token expired**:
   ```bash
   # Solution: Service account tokens are automatically refreshed
   # If using a key file, ensure the file is still valid and readable
   cat $GOOGLE_APPLICATION_CREDENTIALS
   ```

3. **Service account key issues**:
   ```bash
   # Verify the environment variable is set correctly
   echo $GOOGLE_APPLICATION_CREDENTIALS
   
   # Verify the file exists and is readable
   cat $GOOGLE_APPLICATION_CREDENTIALS | jq .client_email
   ```

4. **Wrong service account type**:
   ```bash
   # Ensure the key file is for a service account (not a user account)
   cat $GOOGLE_APPLICATION_CREDENTIALS | jq .type
   # Should output: "service_account"
   ```

### "missing metadata" Error

**Symptoms**: Client gets `rpc error: code = InvalidArgument desc = missing metadata`

**Cause**: Authentication credentials not attached to requests

**Solution**: Ensure `GOOGLE_APPLICATION_CREDENTIALS` is set before running the client

### "could not create ID token source" Error

**Symptoms**: Client gets error about creating ID token source

**Causes and Solutions**:

1. **No credentials set**:
   ```bash
   # Solution: Set GOOGLE_APPLICATION_CREDENTIALS
   export GOOGLE_APPLICATION_CREDENTIALS="$HOME/.config/gcloud/skipctl-key.json"
   ```

2. **Invalid key file**:
   ```bash
   # Verify the JSON file is valid
   cat $GOOGLE_APPLICATION_CREDENTIALS | jq .
   ```

## Best Practices

### Development
- ✅ Use service account keys for consistent authentication
- ✅ Store keys in `~/.config/gcloud/` directory
- ✅ Use self-signed TLS certificates for localhost
- ✅ Set `GOOGLE_APPLICATION_CREDENTIALS` in your shell profile
- ❌ Don't commit service account keys to version control
- ❌ Don't share keys between team members

### Production
- ✅ Use Workload Identity when running in GKE (preferred)
- ✅ Use properly signed TLS certificates from a trusted CA
- ✅ Store service account keys in Secret Manager or similar
- ✅ Rotate service account keys regularly (quarterly recommended)
- ✅ Use separate service accounts for different environments
- ✅ Grant minimal required permissions to service accounts
- ❌ Don't hardcode credentials in environment variables in Dockerfiles
- ❌ Don't use the same key across multiple environments

### CI/CD
- ✅ Use service account keys stored in CI/CD secrets
- ✅ Use short-lived tokens when possible
- ✅ Limit service account permissions to only what's needed
- ✅ Audit service account usage regularly
- ❌ Don't log or print credentials in CI/CD output

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         CLIENT                              │
│                                                             │
│  1. google.DefaultTokenSource()                             │
│     ├─ Checks GOOGLE_APPLICATION_CREDENTIALS env var       │
│     ├─ Checks gcloud user credentials                      │
│     └─ Checks GCE/GKE metadata server                      │
│                                                             │
│  2. Extract ID token from OAuth2 response                  │
│                                                             │
│  3. Attach ID token to gRPC request metadata               │
│     Header: "authorization: Bearer <ID_TOKEN>"             │
│                                                             │
│  4. Establish TLS connection                               │
│     ├─ Localhost: InsecureSkipVerify (self-signed)         │
│     └─ Remote: Validate with system root CAs               │
└─────────────────────────────────────────────────────────────┘
                           │
                           │ Encrypted gRPC/TLS
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                         SERVER                              │
│                                                             │
│  1. TLS handshake                                          │
│     ├─ Server presents certificate                         │
│     └─ Client validates certificate                        │
│                                                             │
│  2. Extract "authorization" header from metadata           │
│                                                             │
│  3. Validate ID token                                      │
│     ├─ Verify signature (Google public keys)               │
│     ├─ Check expiry                                        │
│     └─ Validate claims                                     │
│                                                             │
│  4. Check authentication type                              │
│     ├─ Service account? (email ends with .gserviceaccount) │
│     │  └─ ✅ Allow                                         │
│     └─ User account?                                       │
│        ├─ Check 'hd' claim == kartverket.no               │
│        └─ ✅ Allow if match, ❌ Reject if mismatch         │
│                                                             │
│  5. Log authentication event                               │
│                                                             │
│  6. Inject user context into request                       │
│                                                             │
│  7. Process request                                        │
└─────────────────────────────────────────────────────────────┘
```

## FAQ

**Q: Can I use the same credentials for both server and client?**

A: No. The server needs Vertex AI permissions (AI Platform User role), while the client only needs to authenticate to the server. They can use different service accounts with different permissions.

**Q: Does this work with multiple GCP projects?**

A: Yes. The authentication is based on Google ID tokens which are project-agnostic. The server's GCP project (for Vertex AI) is independent of the authentication project.

**Q: How do I restrict which service accounts can access the server?**

A: Currently, any valid service account from any GCP project can access the server. To restrict access, you would need to add additional validation logic in `validateToken()` to check the email against an allowlist or pattern.

**Q: What happens if my token expires during a long-running operation?**

A: The token is validated at the start of each RPC call. For streaming RPCs, the token is validated once at stream creation. If you need long-running operations, ensure tokens are fresh before starting the operation.
