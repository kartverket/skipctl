package auth

import (
	"context"
	"strings"

	slogcontext "github.com/PumpkinSeed/slog-context"
	"github.com/kartverket/skipctl/pkg/logging"
	"google.golang.org/api/idtoken"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

var (
	errMissingMetadata = status.Errorf(codes.InvalidArgument, "missing metadata")
	errInvalidToken    = status.Errorf(codes.Unauthenticated, "invalid token")
	log                = logging.Logger()
)

func validateToken(ctx context.Context, org string, authorization []string) (string, error) {
	if len(authorization) < 1 {
		return "", errMissingMetadata
	}

	token := strings.TrimPrefix(authorization[0], "Bearer ")
	// we are explicitly not setting an audience as it's random
	payload, err := idtoken.Validate(ctx, token, "")
	if err != nil {
		log.WarnContext(ctx, "error validating token", "error", err)
		return "", errInvalidToken
	}

	email, ok := payload.Claims["email"].(string)
	if !ok || len(email) == 0 {
		log.WarnContext(ctx, "error getting email from token, rejecting further operations", "error", err)
		return "", errInvalidToken
	}

	hd, ok := payload.Claims["hd"].(string)
	if !ok || len(hd) == 0 {
		log.WarnContext(ctx, "claim 'hd' indicating organization not present or empty", "email", email)
		return "", errInvalidToken
	}
	if hd != org {
		log.WarnContext(ctx, "wrong organization present", "wanted", org, "actual", hd)
		return "", errInvalidToken
	}

	return email, nil
}

// ValidADCTokenWithOrg ensures a valid token exists within a request's metadata. The token must
// be scoped to a specific organization. If the token is missing or invalid, the interceptor blocks
// execution of the handler and returns an error. Otherwise, the interceptor invokes the unary
// handler.
func ValidADCTokenWithOrg(idTokenOrg string) func(
	ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			log.WarnContext(ctx, "no metadata present for request")
			return nil, errMissingMetadata
		}
		// The keys within metadata.MD are normalized to lowercase.
		// See: https://godoc.org/google.golang.org/grpc/metadata#New
		auth := md["authorization"]
		email, err := validateToken(ctx, idTokenOrg, auth)
		if err != nil {
			return nil, err
		}

		p, _ := peer.FromContext(ctx)
		userContext := slogcontext.WithValue(ctx, "userInfo", map[string]string{
			"email": email,
			"ip":    p.Addr.String(),
		})
		// Continue execution of handler after ensuring a valid token.
		return handler(userContext, req)
	}
}
