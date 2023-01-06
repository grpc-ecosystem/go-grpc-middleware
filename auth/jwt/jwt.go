package grpc_jwt

import (
	"context"

	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
)

type ContextKey string

// Config defines the config for JWT middleware.
type Config struct {
	// Context key to store user information from the token into context.
	// Optional. Default value "user".
	ContextKey ContextKey

	// Signing key to validate token.
	// This is one of the three options to provide a token validation key.
	// The order of precedence is a user-defined KeyFunc, SigningKeys and SigningKey.
	// Claims will be accepted without verification, if neither user-defined KeyFunc nor SigningKey nor SigningKeys is provided.
	SigningKey interface{}

	// Map of signing keys to validate token with kid field usage.
	// This is one of the three options to provide a token validation key.
	// The order of precedence is a user-defined KeyFunc, SigningKeys and SigningKey.
	// Claims will be accepted without verification, if neither user-defined KeyFunc nor SigningKey nor SigningKeys is provided.
	SigningKeys map[string]interface{}

	// Signing method used to check the token's signing algorithm.
	// Optional. Default value HS256.
	SigningMethod string

	// KeyFunc defines a user-defined function that supplies the public key for a token validation.
	// The function shall take care of verifying the signing algorithm and selecting the proper key.
	// A user-defined KeyFunc can be useful if tokens are issued by an external party.
	// Used by default ParseTokenFunc implementation.
	//
	// When a user-defined KeyFunc is provided, SigningKey, SigningKeys, and SigningMethod are ignored.
	// This is one of the three options to provide a token validation key.
	// The order of precedence is a user-defined KeyFunc, SigningKeys and SigningKey.
	// Claims will be accepted without verification, if neither user-defined KeyFunc nor SigningKey nor SigningKeys is provided.
	// Not used if custom ParseTokenFunc is set or neither user-defined KeyFunc nor SigningKey nor SigningKeys is provided.
	// Default to an internal implementation verifying the signing algorithm and selecting the proper key.
	KeyFunc jwt.Keyfunc

	// AuthScheme to be used in the Authorization header.
	// Optional. Default value "Bearer".
	AuthScheme string

	// ParseTokenFunc defines a user-defined function that parses token from given auth. Returns an error when token
	// parsing fails or parsed token is invalid.
	// Defaults to implementation using `github.com/golang-jwt/jwt` as JWT implementation library
	ParseTokenFunc func(c context.Context, auth string) (interface{}, error)

	// Claims are extendable claims data defining token content. Used by default ParseTokenFunc implementation.
	// Not used if custom ParseTokenFunc is set.
	// Optional. Defaults to function returning jwt.MapClaims
	NewClaimsFunc func(c context.Context) jwt.Claims
}

const (
	// AlgorithmHS256 is token signing algorithm
	AlgorithmHS256    string     = "HS256"
	DefaultContextKey ContextKey = "user"
)

func NewAuthFunc(signingKey interface{}) grpc_auth.AuthFunc {
	return NewAuthFuncWithConfig(Config{SigningKey: signingKey})
}

func NewAuthFuncWithConfig(config Config) grpc_auth.AuthFunc {
	config.setDefaults()
	return func(c context.Context) (context.Context, error) {
		auth, err := grpc_auth.AuthFromMD(c, config.AuthScheme)
		if err != nil {
			return nil, err
		}
		token, err := config.ParseTokenFunc(c, auth)
		if err != nil {
			return nil, err
		}
		newCtx := context.WithValue(c, config.ContextKey, token)
		return newCtx, nil
	}
}

func (config *Config) setDefaults() {
	if config.ContextKey == "" {
		config.ContextKey = DefaultContextKey
	}
	if config.AuthScheme == "" {
		config.AuthScheme = "Bearer"
	}
	if config.SigningMethod == "" {
		config.SigningMethod = AlgorithmHS256
	}
	if config.NewClaimsFunc == nil {
		config.NewClaimsFunc = func(c context.Context) jwt.Claims {
			return jwt.MapClaims{}
		}
	}
	if config.ParseTokenFunc == nil {
		if config.SigningKey == nil && len(config.SigningKeys) == 0 && config.KeyFunc == nil {
			config.ParseTokenFunc = config.defaultParseTokenFuncWithoutVerify
		} else {
			config.ParseTokenFunc = config.defaultParseTokenFunc
		}
	}
	if config.KeyFunc == nil {
		config.KeyFunc = config.defaultKeyFunc
	}
}

func (config *Config) defaultKeyFunc(token *jwt.Token) (interface{}, error) {
	if token.Method.Alg() != config.SigningMethod {
		return nil, status.Errorf(codes.Unauthenticated, "unexpected jwt signing method=%v", token.Header["alg"])
	}
	if len(config.SigningKeys) == 0 {
		return config.SigningKey, nil
	}
	if kid, ok := token.Header["kid"].(string); ok {
		if key, ok := config.SigningKeys[kid]; ok {
			return key, nil
		}
	}
	return nil, status.Errorf(codes.Unauthenticated, "unexpected jwt key id=%v", token.Header["kid"])
}

func (config *Config) defaultParseTokenFunc(c context.Context, auth string) (interface{}, error) {
	token, err := jwt.ParseWithClaims(auth, config.NewClaimsFunc(c), config.KeyFunc)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}
	if !token.Valid {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token")
	}
	return token, nil
}

func (config *Config) defaultParseTokenFuncWithoutVerify(c context.Context, auth string) (interface{}, error) {
	token, _, err := jwt.NewParser().ParseUnverified(auth, config.NewClaimsFunc(c))
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}
	return token, nil
}
