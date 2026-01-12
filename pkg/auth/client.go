package auth

import (
	"context"
	"fmt"

	"google.golang.org/api/idtoken"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

type grpcTokenSource struct {
	oauth.TokenSource
}

// NewADCBackedRPCCredentials creates gRPC credentials using Google Cloud ID tokens.
// This works with service account credentials from GOOGLE_APPLICATION_CREDENTIALS
// or default credentials in GCE/GKE environments.
func NewADCBackedRPCCredentials() (credentials.PerRPCCredentials, error) {
	ctx := context.Background()

	// Service accounts require an audience for ID tokens
	// The server validates the token signature but doesn't check the audience
	audience := "https://skipctl.kartverket.no"

	ts, err := idtoken.NewTokenSource(ctx, audience)
	if err != nil {
		return nil, fmt.Errorf("could not create ID token source: %w (ensure GOOGLE_APPLICATION_CREDENTIALS is set to a service account key file)", err)
	}

	return &grpcTokenSource{TokenSource: oauth.TokenSource{
		TokenSource: ts,
	}}, nil
}
