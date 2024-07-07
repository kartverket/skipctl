package auth

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

type grpcTokenSource struct {
	oauth.TokenSource
}

// idTokenSource is an oauth2.TokenSource that wraps another
// It takes the id_token from TokenSource and passes that on as a bearer token
type idTokenSource struct {
	TokenSource oauth2.TokenSource
}

func (s *idTokenSource) Token() (*oauth2.Token, error) {
	token, err := s.TokenSource.Token()
	if err != nil {
		return nil, err
	}

	idToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("token did not contain an id_token")
	}

	return &oauth2.Token{
		AccessToken: idToken,
		TokenType:   "Bearer",
		Expiry:      token.Expiry,
	}, nil
}

func NewADCBackedRPCCredentials() (credentials.PerRPCCredentials, error) {
	ts, err := google.DefaultTokenSource(context.Background())
	if err != nil {
		return nil, fmt.Errorf("could not get idtoken source: %w", err)
	}
	return &grpcTokenSource{TokenSource: oauth.TokenSource{
		TokenSource: &idTokenSource{TokenSource: ts},
	}}, nil
}
