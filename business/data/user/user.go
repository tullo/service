// Package user contains user related CRUD functionality.
package user

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/dgrijalva/jwt-go/v4"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/tullo/service/business/auth"
	"github.com/tullo/service/business/data"
	"github.com/tullo/service/foundation/database"
	"go.opentelemetry.io/otel/trace"
)

// User manages the set of API's for user access.
type User struct {
	log *log.Logger
	db  *sqlx.DB
}

// New constructs a User for api access.
func New(log *log.Logger, db *sqlx.DB) User {
	return User{
		log: log,
		db:  db,
	}
}

// Create inserts a new user into the database.
func (u User) Create(ctx context.Context, traceID string, nu NewUser, now time.Time) (Info, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.user.create")
	defer span.End()

	hash, err := argon2id.CreateHash(nu.Password, argon2id.DefaultParams)
	if err != nil {
		return Info{}, errors.Wrap(err, "generating password hash")
	}

	usr := Info{
		ID:           uuid.New().String(),
		Name:         nu.Name,
		Email:        nu.Email,
		PasswordHash: hash,
		Roles:        nu.Roles,
		DateCreated:  now.UTC(),
		DateUpdated:  now.UTC(),
	}

	const q = `
	INSERT INTO users
		(user_id, name, email, password_hash, roles, date_created, date_updated)
	VALUES
		(:user_id, :name, :email, :password_hash, :roles, :date_created, :date_updated)`

	u.log.Printf("%s: %s: %s", traceID, "user.Create",
		database.Log(q, usr.ID, usr.Name, usr.Email, "***", usr.Roles, usr.DateCreated, usr.DateUpdated),
	)

	if _, err = u.db.NamedExecContext(ctx, q, usr); err != nil {
		return Info{}, errors.Wrap(err, "inserting user")
	}

	return usr, nil
}

// Update replaces a user document in the database.
func (u User) Update(ctx context.Context, traceID string, claims auth.Claims, userID string, uu UpdateUser, now time.Time) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.user.update")
	defer span.End()

	usr, err := u.QueryByID(ctx, traceID, claims, userID)
	if err != nil {
		return err
	}

	if uu.Name != nil {
		usr.Name = *uu.Name
	}
	if uu.Email != nil {
		usr.Email = *uu.Email
	}
	if uu.Roles != nil {
		usr.Roles = uu.Roles
	}
	if uu.Password != nil {
		hash, err := argon2id.CreateHash(*uu.Password, argon2id.DefaultParams)
		if err != nil {
			return errors.Wrap(err, "generating password hash")
		}
		usr.PasswordHash = hash
	}

	usr.DateUpdated = now

	const q = `
	UPDATE
		users
	SET 
		"name" = :name,
		"email" = :email,
		"roles" = :roles,
		"password_hash" = :password_hash,
		"date_updated" = :date_updated
	WHERE
		user_id = :user_id`

	u.log.Printf("%s: %s: %s", traceID, "user.Update",
		database.Log(q, usr.ID, usr.Name, usr.Email, usr.Roles, "***", usr.DateCreated, usr.DateUpdated),
	)

	if _, err = u.db.NamedExecContext(ctx, q, usr); err != nil {
		return errors.Wrap(err, "updating user")
	}

	return nil
}

// Delete removes a user from the database.
func (u User) Delete(ctx context.Context, traceID string, claims auth.Claims, userID string) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.user.delete")
	defer span.End()

	if _, err := uuid.Parse(userID); err != nil {
		return data.ErrInvalidID
	}

	if !claims.Authorized(auth.RoleAdmin) { // If you are not an admin
		if claims.Subject != userID { // and looking to delete someone other than yourself.
			return data.ErrForbidden
		}
	}

	const q = `
	DELETE FROM
		users
	WHERE
		user_id = :user_id`

	u.log.Printf("%s: %s: %s", traceID, "user.Delete",
		database.Log(q, userID),
	)

	usr := Info{ID: userID}
	if _, err := u.db.NamedExecContext(ctx, q, usr); err != nil {
		return errors.Wrapf(err, "deleting user %s", userID)
	}

	return nil
}

// Query retrieves a list of existing users from the database.
func (u User) Query(ctx context.Context, traceID string, pageNumber int, rowsPerPage int) ([]Info, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.user.query")
	defer span.End()

	const q = `
	SELECT
		*
	FROM
		users
	ORDER BY
		user_id
	OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY`

	page := struct {
		Offset      int `db:"offset"`
		RowsPerPage int `db:"rows_per_page"`
	}{
		Offset:      (pageNumber - 1) * rowsPerPage,
		RowsPerPage: rowsPerPage,
	}

	u.log.Printf("%s: %s: %s", traceID, "user.Query",
		database.Log(q, page),
	)

	ns, err := u.db.PrepareNamedContext(ctx, q)
	if err != nil {
		return nil, errors.Wrap(err, "prepare named context")
	}
	defer ns.Close()

	users := make([]Info, 0, page.RowsPerPage)
	if err = ns.SelectContext(ctx, &users, page); err != nil {
		return nil, errors.Wrap(err, "query products")
	}

	return users, nil
}

// QueryByID gets the specified user from the database.
func (u User) QueryByID(ctx context.Context, traceID string, claims auth.Claims, userID string) (Info, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.user.querybyid")
	defer span.End()

	if _, err := uuid.Parse(userID); err != nil {
		return Info{}, data.ErrInvalidID
	}

	if !claims.Authorized(auth.RoleAdmin) { // If you are not an admin
		if claims.Subject != userID { // and looking to retrieve someone other than yourself.
			return Info{}, data.ErrForbidden
		}
	}

	const q = `
	SELECT
		*
	FROM
		users
	WHERE 
		user_id = :user_id`

	u.log.Printf("%s: %s: %s", traceID, "user.QueryByID",
		database.Log(q, userID),
	)

	ns, err := u.db.PrepareNamedContext(ctx, q)
	if err != nil {
		return Info{}, errors.Wrap(err, "prepare named context")
	}
	defer ns.Close()

	var usr Info
	if err := ns.GetContext(ctx, &usr, Info{ID: userID}); err != nil {
		if err == sql.ErrNoRows {
			return Info{}, data.ErrNotFound
		}
		return Info{}, errors.Wrapf(err, "selecting user %q", userID)
	}

	return usr, nil
}

// QueryByEmail gets the specified user from the database by email.
func (u User) QueryByEmail(ctx context.Context, traceID string, claims auth.Claims, email string) (Info, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.user.querybyemail")
	defer span.End()

	const q = `
	SELECT
		*
	FROM
		users
	WHERE
		email = :email`

	u.log.Printf("%s: %s: %s", traceID, "user.QueryByEmail",
		database.Log(q, email),
	)

	ns, err := u.db.PrepareNamedContext(ctx, q)
	if err != nil {
		return Info{}, errors.Wrap(err, "prepare named context")
	}
	defer ns.Close()

	var usr Info
	if err := ns.GetContext(ctx, &usr, Info{Email: email}); err != nil {
		if err == sql.ErrNoRows {
			return Info{}, data.ErrNotFound
		}
		return Info{}, errors.Wrapf(err, "selecting user %q", email)
	}

	if !claims.Authorized(auth.RoleAdmin) { // If you are not an admin
		if claims.Subject != usr.ID { // and looking to retrieve someone other than yourself.
			return Info{}, data.ErrForbidden
		}
	}

	return usr, nil
}

// Authenticate finds a user by their email and verifies their password. On
// success it returns claims representing this user. The claims can be used to
// generate a token for future authentication.
func (u User) Authenticate(ctx context.Context, traceID string, now time.Time, email, password string) (auth.Claims, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.user.authenticate")
	defer span.End()

	const q = `
	SELECT
		*
	FROM
		users
	WHERE
		email = :email`

	u.log.Printf("%s: %s: %s", traceID, "user.Authenticate",
		database.Log(q, email),
	)

	ns, err := u.db.PrepareNamedContext(ctx, q)
	if err != nil {
		return auth.Claims{}, errors.Wrap(err, "prepare named context")
	}
	defer ns.Close()

	var usr Info
	if err := ns.GetContext(ctx, &usr, Info{Email: email}); err != nil {

		// Normally we would return ErrNotFound in this scenario but we do not want
		// to leak to an unauthenticated user which emails are in the system.
		if err == sql.ErrNoRows {
			return auth.Claims{}, data.ErrAuthenticationFailure
		}

		return auth.Claims{}, errors.Wrapf(err, "selecting user %q", email)
	}

	// Compare the provided password with the saved hash. Use the bcrypt
	// comparison function so it is cryptographically secure.
	if match, err := argon2id.ComparePasswordAndHash(password, usr.PasswordHash); err != nil || !match {
		return auth.Claims{}, data.ErrAuthenticationFailure
	}

	// If we are this far the request is valid. Create some claims for the user
	// and generate their token.
	claims := auth.Claims{
		StandardClaims: jwt.StandardClaims{
			Issuer:    "service project",
			Audience:  jwt.ClaimStrings{"students"},
			Subject:   usr.ID,
			ExpiresAt: jwt.At(now.Add(time.Hour)),
			IssuedAt:  jwt.At(now),
		},
		Roles: usr.Roles,
	}

	return claims, nil
}
