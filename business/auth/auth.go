package auth

import (
	"crypto/rsa"
	"fmt"
	"sync"

	"github.com/dgrijalva/jwt-go/v4"
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
func (c Claims) Valid(h *jwt.ValidationHelper) error {
	for _, r := range c.Roles {
		switch r {
		case RoleAdmin, RoleUser: // Role is valid.
		default:
			return fmt.Errorf("invalid role %q", r)
		}
	}
	if err := c.StandardClaims.Valid(h); err != nil {
		return errors.Wrap(err, "validating standard claims")
	}
	return nil
}

// Authorized returns true if the claims has at least one of the provided roles.
func (c Claims) Authorized(roles ...string) bool {
	for _, has := range c.Roles {
		for _, want := range roles {
			if has == want {
				return true
			}
		}
	}
	return false
}

// Keys represents an in memory store of keys.
type Keys map[string]*rsa.PrivateKey

// PublicKeyLookup defines the signature of a function to lookup public keys.
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
type PublicKeyLookup func(kid string) (*rsa.PublicKey, error)

// Auth is used to authenticate clients. It can generate a token for a
// set of user claims and recreate the claims by parsing the token.
type Auth struct {
	algorithm string
	keyFunc   func(t *jwt.Token) (interface{}, error)
	keys      Keys
	method    jwt.SigningMethod
	mu        sync.RWMutex
	parser    *jwt.Parser
}

// New creates an *Auth for use. It will error if:
// - The private key is nil.
// - The public key func is nil.
// - The key ID is blank.
// - The specified algorithm is unsupported.
func New(algorithm string, lookup PublicKeyLookup) (*Auth, error) {
	method := jwt.GetSigningMethod(algorithm)
	if method == nil {
		return nil, errors.Errorf("unknown algorithm %v", algorithm)
	}

	if lookup == nil {
		return nil, errors.New("public key lookup function cannot be nil")
	}

	// keyFunc is a function that returns the public key for validating a token.
	// We use the parsed (but unverified) token to find the key id. That KID is
	// passed to our KeyFunc to find the public key to use for verification.
	keyFunc := func(t *jwt.Token) (interface{}, error) {
		kid, ok := t.Header["kid"]
		if !ok {
			return nil, errors.New("missing key id (kid) in token header")
		}
		publicKID, ok := kid.(string)
		if !ok {
			return nil, errors.New("user token key id (kid) must be string")
		}
		return lookup(publicKID)
	}

	// Create the token parser to use. The algorithm used to sign the JWT must be
	// validated to avoid a critical vulnerability:
	// https://auth0.com/blog/critical-vulnerabilities-in-json-web-token-libraries/
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{algorithm}),
		jwt.WithAudience("students"),
	)
	a := Auth{
		algorithm: algorithm,
		keyFunc:   keyFunc,
		keys:      make(map[string]*rsa.PrivateKey),
		method:    method,
		parser:    parser,
	}

	return &a, nil
}

// AddKey adds a kid and private key combination to our local store. It returns
// the updated size if the key store.
func (a *Auth) AddKey(kid string, privateKey *rsa.PrivateKey) int {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.keys[kid] = privateKey
	return len(a.keys)
}

// RemoveKey removes a kid and private key combination from our local store. It
// returns the updated size if the key store.
func (a *Auth) RemoveKey(kid string) int {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.keys, kid)
	return len(a.keys)
}

// GenerateToken generates a signed JWT token string representing the user Claims.
func (a *Auth) GenerateToken(kid string, claims Claims) (string, error) {
	token := jwt.NewWithClaims(a.method, claims)
	token.Header["kid"] = kid

	var privateKey *rsa.PrivateKey
	a.mu.RLock()
	{
		var ok bool
		privateKey, ok = a.keys[kid]
		if !ok {
			return "", errors.New("kid lookup failed")
		}
	}
	a.mu.RUnlock()

	str, err := token.SignedString(privateKey)
	if err != nil {
		return "", errors.Wrap(err, "signing token")
	}

	return str, nil
}

// ValidateToken recreates the Claims that were used to generate a token. It
// verifies that the token was signed using our key.
func (a *Auth) ValidateToken(tokenStr string) (Claims, error) {

	var claims Claims
	token, err := a.parser.ParseWithClaims(tokenStr, &claims, a.keyFunc)
	if err != nil {
		return Claims{}, errors.Wrap(err, "parsing token")
	}

	if !token.Valid {
		return Claims{}, errors.New("invalid token")
	}

	return claims, nil
}
