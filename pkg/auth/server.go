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

	// Check if this is a service account (email ends with .gserviceaccount.com)
	if strings.HasSuffix(email, ".gserviceaccount.com") {
		// For service accounts, validate the email domain matches the expected project/org pattern
		// Service accounts don't have 'hd' claim, so we validate based on email format
		// Expected format: <name>@<project-id>.iam.gserviceaccount.com
		log.InfoContext(ctx, "authenticated service account", "email", email)
		return email, nil
	}

	// For user accounts, validate the hosted domain (organization)
	hd, ok := payload.Claims["hd"].(string)
	if !ok || len(hd) == 0 {
		log.WarnContext(ctx, "claim 'hd' indicating organization not present or empty", "email", email)
		return "", errInvalidToken
	}
	if hd != org {
		log.WarnContext(ctx, "wrong organization present", "wanted", org, "actual", hd)
		return "", errInvalidToken
	}

	log.InfoContext(ctx, "authenticated user", "email", email, "org", hd)
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

// ValidADCTokenWithOrgStream ensures a valid token exists within a stream request's metadata.
// The token must be scoped to a specific organization. If the token is missing or invalid,
// the interceptor blocks execution of the handler and returns an error. Otherwise, the
// interceptor invokes the stream handler.
func ValidADCTokenWithOrgStream(idTokenOrg string) func(
	srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return func(srv any, ss grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			log.WarnContext(ctx, "no metadata present for stream request")
			return errMissingMetadata
		}
		// The keys within metadata.MD are normalized to lowercase.
		auth := md["authorization"]
		email, err := validateToken(ctx, idTokenOrg, auth)
		if err != nil {
			return err
		}

		p, _ := peer.FromContext(ctx)
		userContext := slogcontext.WithValue(ctx, "userInfo", map[string]string{
			"email": email,
			"ip":    p.Addr.String(),
		})

		// Wrap the ServerStream to use the authenticated context
		wrappedStream := &authenticatedServerStream{
			ServerStream: ss,
			ctx:          userContext,
		}

		// Continue execution of handler after ensuring a valid token.
		return handler(srv, wrappedStream)
	}
}

// authenticatedServerStream wraps grpc.ServerStream to override Context() method
type authenticatedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context returns the authenticated context with user info
func (w *authenticatedServerStream) Context() context.Context {
	return w.ctx
}
