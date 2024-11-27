package main

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
)

// Custom claims to get the permissions from our access token
type CustomClaims struct {
	Permissions  []string `json:"permissions"`
	Roles        []string `json:"https://api.sambego.tech/roles"` // custom claims should be prefixed
	ShouldReject bool     `json:"shouldReject,omitempty"`
}

func (c *CustomClaims) Validate(ctx context.Context) error {
	return nil
}

func EnsureValidAccessToken() func(next http.Handler) http.Handler {
	// Get the issuer URL from .env
	issuerURL, err := url.Parse(os.Getenv("AUTH0_DOMAIN"))

	if err != nil {
		log.Fatalf("Failed to parse the issuer url: %v", err)
	}

	// Create a new caching provider to cache the JWK
	provider := jwks.NewCachingProvider(issuerURL, 5*time.Minute)

	// Create a new JWT validator
	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL.String(),
		[]string{os.Getenv("AUTH0_AUDIENCE")},
		validator.WithCustomClaims(
			func() validator.CustomClaims {
				return &CustomClaims{}
			},
		),
	)

	// Handle errors
	if err != nil {
		log.Fatalf("Failed to set up the jwt validator")
	}

	// Create the error handler for the middleware
	errorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Encountered error while validating JWT: %v", err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Failed to validate JWT."}`))
	}

	// Create the JWT middleware
	middleware := jwtmiddleware.New(
		jwtValidator.ValidateToken,
		jwtmiddleware.WithErrorHandler(errorHandler),
	)

	return func(next http.Handler) http.Handler {
		return middleware.CheckJWT(next)
	}
}

func NewEnsureValidAccessToken(next http.Handler) http.Handler {
	return EnsureValidAccessToken()(next)
}
