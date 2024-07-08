package auth

import (
	"context"
	"strings"

	"google.golang.org/api/idtoken"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	errMissingMetadata = status.Errorf(codes.InvalidArgument, "missing metadata")
	errInvalidToken    = status.Errorf(codes.Unauthenticated, "invalid token")
)

type contextKey string

func validateToken(ctx context.Context, org string, authorization []string) (string, error) {
	if len(authorization) < 1 {
		return "", errMissingMetadata
	}

	token := strings.TrimPrefix(authorization[0], "Bearer ")
	// we are explicitly not setting an audience as it's random
	payload, err := idtoken.Validate(ctx, token, "")
	if err != nil {
		log.ErrorContext(ctx, "error validating token", "error", err)
		return "", errInvalidToken
	}

	hd, ok := payload.Claims["hd"].(string)
	if !ok || len(hd) == 0 {
		log.WarnContext(ctx, "claim 'hd' indicating organization not present or empty", "subject", payload.Subject)
		return "", errInvalidToken
	}
	if hd != org {
		log.WarnContext(ctx, "wrong organization present", "wanted", org, "actual", hd)
		return "", errInvalidToken
	}

	return payload.Claims["email"].(string), nil
}

// ValidADCTokenWithOrg ensures a valid token exists within a request's metadata. The token must
// be scoped to a specific organization. If the token is missing or invalid, the interceptor blocks
// execution of the handler and returns an error. Otherwise, the interceptor invokes the unary
// handler.
func ValidADCTokenWithOrg(idTokenOrg string) func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, errMissingMetadata
		}
		// The keys within metadata.MD are normalized to lowercase.
		// See: https://godoc.org/google.golang.org/grpc/metadata#New
		auth := md["authorization"]
		email, err := validateToken(ctx, idTokenOrg, auth)
		if err != nil {
			return nil, err
		}
		log.InfoContext(ctx, "I am authz", "email", email)
		newCtx := context.WithValue(ctx, contextKey("email"), email)
		// Continue execution of handler after ensuring a valid token.
		return handler(newCtx, req)
	}
}
