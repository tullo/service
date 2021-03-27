package auth_test

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go/v4"
	"github.com/tullo/service/business/auth"
)

const publicTestKID = "54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"

// Success and failure markers.
const (
	success = "\u2713"
	failed  = "\u2717"
)

func TestAuth(t *testing.T) {
	t.Log("Given the need to be able to AuthN and AuthZ access.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a single user.", testID)
		{
			privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to create a private key: %v", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to create a private key.", success, testID)

			lookup := func(kid string) (*rsa.PublicKey, error) {
				switch kid {
				case publicTestKID:
					return &privateKey.PublicKey, nil
				}
				return nil, fmt.Errorf("no public key found for the specified kid: %s", kid)
			}

			a, err := auth.New("RS256", lookup)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to create an authenticator: %v", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to create an authenticator.", success, testID)

			keys := a.AddKey(publicTestKID, privateKey)
			if keys != 1 {
				t.Fatalf("\t%s\tTest %d:\tShould be able to add a [kid:private-key] combination: keys in store (%d)", failed, testID, keys)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to add a [kid:private-key] combination.", success, testID)

			kid := "a-signing-key-id"
			keys = a.AddKey(kid, privateKey)
			if keys != 2 {
				t.Fatalf("\t%s\tTest %d:\tShould be able to add a [kid:private-key] combination: keys in store (%d)", failed, testID, keys)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to add a [kid:private-key] combination.", success, testID)

			keys = a.RemoveKey(kid)
			if keys != 1 {
				t.Fatalf("\t%s\tTest %d:\tShould be able to remove a [kid:private-key] combination: keys in store (%d)", failed, testID, keys)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to remove a [kid:private-key] combination.", success, testID)

			claims := auth.Claims{
				StandardClaims: jwt.StandardClaims{
					Issuer:    "service project",
					Subject:   "5cf37266-3473-4006-984f-9325122678b7",
					Audience:  jwt.ClaimStrings{"students"},
					ExpiresAt: jwt.At(time.Now().Add(8760 * time.Hour)),
					IssuedAt:  jwt.Now(),
				},
				Roles: []string{auth.RoleAdmin},
			}

			token, err := a.GenerateToken(publicTestKID, claims)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to generate a JWT: %v", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to generate a JWT.", success, testID)

			parsedClaims, err := a.ValidateToken(token)
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
