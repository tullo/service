package auth_test

import (
	"crypto/rsa"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"github.com/tullo/service/business/auth"
)

func TestAuthenticator(t *testing.T) {

	// Parse the private key used to generate the token.
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(privateRSAKey))
	if err != nil {
		t.Fatal(err)
	}

	// Create a key lookup function that returns the public key for the test KID.
	keyLookupFunc := func(publicKID string) (*rsa.PublicKey, error) {
		if publicKID != keyID {
			return nil, errors.New("no public key found")
		}
		return &privateKey.PublicKey, nil
	}
	a, err := auth.New(privateKey, keyID, "RS256", keyLookupFunc)
	if err != nil {
		t.Fatal(err)
	}

	claims := auth.Claims{
		StandardClaims: jwt.StandardClaims{
			Issuer:    "service project",
			Subject:   "0x01",
			Audience:  "students",
			ExpiresAt: time.Now().Add(8760 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		Roles: []string{"ADMIN"},
	}

	token, err := a.GenerateToken(claims)
	if err != nil {
		t.Fatal(err)
	}

	parsedClaims, err := a.ValidateToken(token)
	if err != nil {
		t.Fatal(err)
	}

	if exp, got := len(claims.Roles), len(parsedClaims.Roles); exp != got {
		t.Fatalf("expected %v roles, got %v", exp, got)
	}
	if exp, got := claims.Roles[0], parsedClaims.Roles[0]; exp != got {
		t.Fatalf("expected roles[0] == %v, got %v", exp, got)
	}
}

// The key id we would have generated for the keys below.
// The key id represents the public key in the public key store.
const keyID = "54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"

// Output of:
// openssl genpkey -algorithm RSA -out private.pem -pkeyopt rsa_keygen_bits:2048
const privateRSAKey = `-----BEGIN PRIVATE KEY-----
MIIEwAIBADANBgkqhkiG9w0BAQEFAASCBKowggSmAgEAAoIBAQDdiBDU4jqRYuHl
yBmo5dWB1j9aeDrXzUTJbRKlgo+DWDQzIzJQvackvRu8/f7B5cseoqmeJcmBu6pc
4DmQ+puGNHxzCyYVFSMwRtHBZvfWS3P+UqIXCKRAX/NZbLkUEeqPnn5WXjA+YXKk
sfniE0xDH8W22o0OXHOzRhDWORjNTulpMpLv8tKnnLKh2Y/kCL/4vo0SZ+RWh8F9
4+JTZx/47RHWb6fkxkikyTO3zO3efIkrKjfRx2CwFwO2rQ/3T04GQB/Lgr5lfJQU
iofvvVYuj2xBJao+3t9Ir0OeSbw1T5Rz03VLtN8SZhvaxWaBfwkUuUNL1glJO+Yd
LkMxGS0zAgMBAAECggEBAKM6m7RQUPlJE8u8qfOCDdSSKbIefrT9wZ5tKN0dG2Oa
/TNkzrEhXOO8F5Ek0a7LA+Q51KL7ksNtpLS0XpZNoYS8bapS36ePIJN0yx8nIJwc
koYlGtu/+U6ZpHQSoTiBjwRtswcudXuxT8i8frOupnWbFpKJ7H9Vbcb9bHB8N6Mm
D63wSBR08ZMrZXheKHQCQcxSQ2ZQZ+X3LBIOdXZH1aaptU2KpMEU5oyxXPShTVMg
0f748yU2njXCF0ZABEanXgp13egr/MPqHwnS/h0PH45bNy3IgFtMEHEouQFsAzoS
qNe8/9WnrpY87UdSZMnzF/IAXV0bmollDnqfM8/EqxkCgYEA96ThXYGzAK5RKNqp
RqVdRVA0UTT48sJvrxLMuHpyUzg6cl8FZE5rrNxFbouxvyN192Ctv1q8yfv4/HfM
KpmtEjt3fYtITHVXII6O3qNaRoIEPwKT4eK/ar+JO59vI0YvweXvDH5TkS9aiFr+
pPGf3a7EbE24BKhgiI8eT6K0VuUCgYEA5QGg11ZVoUut4ERAPouwuSdWwNe0HYqJ
A1m5vTvF5ghUHAb023lrr7Psq9DPJQQe7GzPfXafsat9hGenyqiyxo1gwClIyoEH
fOg753kdHcy60VVzumsPXece3OOSnd0rRMgfsSsclgYO7z0g9YZPAjt2w9NVw6uN
UDqX3eO2WjcCgYEA015eoNHv99fRG96udsbz+hI/5UQibAl7C+Iu7BJO/CrU8AOc
dYXdr5f+hyEioDLjIDbbdaU71+aCGPMjRwUNzK8HCRfVqLTKndYvqWWhyuZ0O1e2
4ykHGlTLDCHD2Uaxwny/8VjteNEDI7kO+bfmLG9b5djcBNW2Nzh4tZ348OUCgYEA
vIrTppbhF1QciqkGj7govrgBt/GfzDaTyZtkzcTZkSNIRG8Bx3S3UUh8UZUwBpTW
9OY9ClnQ7tF3HLzOq46q6cfaYTtcP8Vtqcv2DgRsEW3OXazSBChC1ZgEk+4Vdz1x
c0akuRP6jBXe099rNFno0LiudlmXoeqrBOPIxxnEt48CgYEAxNZBc/GKiHXz/ZRi
IZtRT5rRRof7TEiDxSKOXHSG7HhIRDCrpwn4Dfi+GWNHIwsIlom8FzZTSHAN6pqP
E8Imrlt3vuxnUE1UMkhDXrlhrxslRXU9enynVghAcSrg6ijs8KuN/9RB/I7H03cT
77mx9eHMcYcRUciY5C8AOaArmMA=
-----END PRIVATE KEY-----`

// How to generate a public key PEM file.
// openssl rsa -pubout -in private.pem -out public.pem