package test

import (
	"log"

	"cloud.google.com/go/auth"
	"cloud.google.com/go/auth/credentials"
)

var tokenProvider auth.TokenProvider

func init() {
	creds, err := credentials.DetectDefault(&credentials.DetectOptions{
		Scopes: []string{"https://www.googleapis.com/auth/userinfo.email"},
	})
	if err != nil {
		log.Fatal(err)
	}

	tokenProvider = creds.TokenProvider
}
