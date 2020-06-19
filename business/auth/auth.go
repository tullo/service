package auth

import (
	"crypto/rsa"
	"fmt"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

// These are the expected values for Claims.Roles.
const (
	RoleAdmin = "ADMIN"
	RoleUser  = "USER"
)

// ctxKey represents the type of value for the context key.
type ctxKey int

// Key is used to store/retrieve a Claims value from a context.Context.
const Key ctxKey = 1

// Claims represents the authorization claims transmitted via a JWT.
type Claims struct {
	Roles []string `json:"roles"`
	jwt.StandardClaims
}

// Valid is called during the parsing of a token.
func (c Claims) Valid() error {
	for _, r := range c.Roles {
		switch r {
		case RoleAdmin, RoleUser: // Role is valid.
		default:
			return fmt.Errorf("invalid role %q", r)
		}
	}
	if err := c.StandardClaims.Valid(); err != nil {
		return errors.Wrap(err, "validating standard claims")
	}
	return nil
}

// HasRole returns true if the claims has at least one of the provided roles.
func (c Claims) HasRole(roles ...string) bool {
	for _, has := range c.Roles {
		for _, want := range roles {
			if has == want {
				return true
			}
		}
	}
	return false
}

// KeyLookupFunc defines the signature of a function to lookup public keys.
//
// In a production system, a key id (KID) is used to retrieve the correct
// public key to parse a JWT for auth and claims. A key lookup function is
// provided to perform the task of retrieving a KID for a given public key.
//
// A key lookup function is required for creating an Authenticator.
//
// * Private keys should be rotated. During the transition period, tokens
// signed with the old and new keys can coexist by looking up the correct
// public key by KID.
//
// * KID to public key resolution is usually accomplished via a public JWKS
// endpoint. See https://auth0.com/docs/jwks for more details.
type KeyLookupFunc func(kid string) (*rsa.PublicKey, error)

// Auth is used to authenticate clients. It can generate a token for a
// set of user claims and recreate the claims by parsing the token.
type Auth struct {
	privateKey       *rsa.PrivateKey
	activeKID        string
	algorithm        string
	pubKeyLookupFunc KeyLookupFunc
	parser           *jwt.Parser
}

// New creates an *Auth for use. It will error if:
// - The private key is nil.
// - The public key func is nil.
// - The key ID is blank.
// - The specified algorithm is unsupported.
func New(privateKey *rsa.PrivateKey, activeKID, algorithm string, publicKeyLookupFunc KeyLookupFunc) (*Auth, error) {
	if privateKey == nil {
		return nil, errors.New("private key cannot be nil")
	}
	if activeKID == "" {
		return nil, errors.New("active kid cannot be blank")
	}
	if jwt.GetSigningMethod(algorithm) == nil {
		return nil, errors.Errorf("unknown algorithm %v", algorithm)
	}
	if publicKeyLookupFunc == nil {
		return nil, errors.New("public key function cannot be nil")
	}

	// Create the token parser to use. The algorithm used to sign the JWT must be
	// validated to avoid a critical vulnerability:
	// https://auth0.com/blog/critical-vulnerabilities-in-json-web-token-libraries/
	parser := jwt.Parser{
		ValidMethods: []string{algorithm},
	}

	a := Auth{
		privateKey:       privateKey,
		activeKID:        activeKID,
		algorithm:        algorithm,
		pubKeyLookupFunc: publicKeyLookupFunc,
		parser:           &parser,
	}

	return &a, nil
}

// GenerateToken generates a signed JWT token string representing the user Claims.
func (a *Auth) GenerateToken(claims Claims) (string, error) {
	method := jwt.GetSigningMethod(a.algorithm)

	tkn := jwt.NewWithClaims(method, claims)
	tkn.Header["kid"] = a.activeKID

	str, err := tkn.SignedString(a.privateKey)
	if err != nil {
		return "", errors.Wrap(err, "signing token")
	}

	return str, nil
}

// ValidateToken recreates the Claims that were used to generate a token. It
// verifies that the token was signed using our key.
func (a *Auth) ValidateToken(tokenStr string) (Claims, error) {

	// keyFunc is a function that returns the public key for validating a token.
	// We use the parsed (but unverified) token to find the key id. That ID is
	// passed to our KeyFunc to find the public key to use for verification.
	keyFunc := func(t *jwt.Token) (interface{}, error) {
		kid, ok := t.Header["kid"]
		if !ok {
			return nil, errors.New("missing key id (kid) in token header")
		}
		userKID, ok := kid.(string)
		if !ok {
			return nil, errors.New("user token key id (kid) must be string")
		}
		return a.pubKeyLookupFunc(userKID)
	}

	var claims Claims
	token, err := a.parser.ParseWithClaims(tokenStr, &claims, keyFunc)
	if err != nil {
		return Claims{}, errors.Wrap(err, "parsing token")
	}

	if !token.Valid {
		return Claims{}, errors.New("invalid token")
	}

	return claims, nil
}
