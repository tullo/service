package commands

import (
	"context"
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/dgrijalva/jwt-go/v4"
	"github.com/pkg/errors"
	"github.com/tullo/service/business/auth"
	"github.com/tullo/service/business/data/user"
	"github.com/tullo/service/foundation/database"
	"github.com/tullo/service/foundation/keystore"
)

// TokenGen generates a JWT for the specified user.
func TokenGen(traceID string, log *log.Logger, cfg database.Config, userID string, privateKeyFile string, algorithm string) error {
	if userID == "" || privateKeyFile == "" || algorithm == "" {
		fmt.Println("help: tokengen <id> <private_key_file> <algorithm>")
		fmt.Println("algorithm: RS256, HS256")
		return ErrHelp
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := database.Connect(ctx, cfg)
	if err != nil {
		return errors.Wrap(err, "connect database")
	}
	defer db.Close()

	u := user.NewStore(log, db)

	// The call to retrieve a user requires an Admin role by the caller.
	claims := auth.Claims{
		StandardClaims: jwt.StandardClaims{
			Subject: userID,
		},
		Roles: []string{auth.RoleAdmin},
	}

	user, err := u.QueryByID(ctx, traceID, claims, userID)
	if err != nil {
		return errors.Wrap(err, "retrieve user")
	}

	privatePEM, err := ioutil.ReadFile(privateKeyFile)
	if err != nil {
		return errors.Wrap(err, "reading PEM private key file")
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privatePEM)
	if err != nil {
		return errors.Wrap(err, "parsing PEM into private key")
	}

	// In a production system, a key id (KID) is used to retrieve the correct
	// public key to parse a JWT for auth and claims. A key store is provided
	// to the auth package for storage and lookup purpose. This id will be
	// assigned to the private key just constructed.
	keyID := "54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"

	// An authenticator maintains the state required to handle JWT processing.
	// It requires a keystore to lookup private and public keys based on a key
	// id. There is a keystore implementation in the project.
	keyPair := map[string]*rsa.PrivateKey{keyID: privateKey}
	keyStore := keystore.NewMap(keyPair)
	a, err := auth.New(algorithm, keyStore)
	if err != nil {
		return errors.Wrap(err, "constructing authenticator")
	}

	// Generating a token requires defining a set of claims. In this applications
	// case, we only care about defining the subject and the user in question and
	// the roles they have on the database. This token will expire in a year.
	//
	// iss (issuer): Issuer of the JWT
	// sub (subject): Subject of the JWT (the user)
	// aud (audience): Recipient for which the JWT is intended
	// exp (expiration time): Time after which the JWT expires
	// nbf (not before time): Time before which the JWT must not be accepted for processing
	// iat (issued at time): Time at which the JWT was issued; can be used to determine age of the JWT
	// jti (JWT ID): Unique identifier; can be used to prevent the JWT from being replayed (allows a token to be used only once)
	claims = auth.Claims{
		StandardClaims: jwt.StandardClaims{
			Issuer:    "service project",
			Subject:   user.ID,
			Audience:  jwt.ClaimStrings{"students"},
			ExpiresAt: jwt.At(time.Now().Add(8760 * time.Hour)),
			IssuedAt:  jwt.Now(),
		},
		Roles: user.Roles,
	}

	// This will generate a JWT with the claims embedded in them. The database
	// with need to be configured with the information found in the public key
	// file to validate these claims. Dgraph does not support key rotate at
	// this time.
	token, err := a.GenerateToken(keyID, claims)
	if err != nil {
		return errors.Wrap(err, "generating token")
	}

	fmt.Printf("-----BEGIN TOKEN-----\n%s\n-----END TOKEN-----\n", token)
	return nil
}
