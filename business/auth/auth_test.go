package auth_test

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"github.com/tullo/service/business/auth"
)

const publicTestKID = "54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"

// Success and failure markers.
const (
	success = "\u2713"
	failed  = "\u2717"
)

func TestAuth(t *testing.T) {
	t.Log("Given the need to be able to authenticate and authorize access.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a single user.", testID)
		{
			// Parse the private key used to generate the token.
			privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to create a private key: %v", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to create a private key.", success, testID)

			// Create a key lookup function that returns the public key for the test KID.
			lookup := func(kid string) (*rsa.PublicKey, error) {
				if kid != publicTestKID {
					return nil, errors.New("no public key found")
				}
				return &privateKey.PublicKey, nil
			}

			authenticator, err := auth.New("RS256", lookup, publicTestKID, privateKey)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to create an authenticator: %v", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to create an authenticator.", success, testID)

			claims := auth.Claims{
				StandardClaims: jwt.StandardClaims{
					Issuer:    "service project",
					Subject:   "5cf37266-3473-4006-984f-9325122678b7",
					Audience:  "students",
					ExpiresAt: time.Now().Add(8760 * time.Hour).Unix(),
					IssuedAt:  time.Now().Unix(),
				},
				Roles: []string{auth.RoleAdmin},
			}

			token, err := authenticator.GenerateToken(claims)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to generate a JWT: %v", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to generate a JWT.", success, testID)

			parsedClaims, err := authenticator.ValidateToken(token)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to parse the claims: %v", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to parse the claims.", success, testID)

			if exp, got := len(claims.Roles), len(parsedClaims.Roles); exp != got {
				t.Logf("\t\tTest %d:\texp: %d", testID, exp)
				t.Logf("\t\tTest %d:\tgot: %d", testID, got)
				t.Fatalf("\t%s\tTest %d:\tShould have the expected number of roles: %v", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould have the expected number of roles.", success, testID)

			if exp, got := claims.Roles[0], parsedClaims.Roles[0]; exp != got {
				t.Logf("\t\tTest %d:\texp: %v", testID, exp)
				t.Logf("\t\tTest %d:\tgot: %v", testID, got)
				t.Fatalf("\t%s\tTest %d:\tShould have the expected roles: %v", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould have the expected roles.", success, testID)
		}
	}
}
