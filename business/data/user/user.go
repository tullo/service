// Package user contains user related CRUD functionality.
package user

import (
	"context"
	"log"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/dgrijalva/jwt-go/v4"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/pkg/errors"
	"github.com/tullo/service/business/auth"
	"github.com/tullo/service/business/data"
	"github.com/tullo/service/foundation/database"
	"go.opentelemetry.io/otel/trace"
)

// https://www.postgresql.org/docs/current/errcodes-appendix.html
const uniqueViolation = "23505"

// Store manages the set of API's for user access.
type Store struct {
	log *log.Logger
	db  *database.DB
}

// NewStore constructs a Store for api access.
func NewStore(log *log.Logger, db *database.DB) Store {
	return Store{
		log: log,
		db:  db,
	}
}

// Create inserts a new user into the database.
func (s Store) Create(ctx context.Context, traceID string, nu NewUser, now time.Time) (Info, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.user.create")
	defer span.End()

	conn, err := s.db.Acquire(ctx)
	if err != nil {
		return Info{}, errors.Wrap(err, "acquire db connection")
	}
	defer conn.Release()

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
		($1, $2, $3, $4, $5, $6, $7)`

	if _, err := conn.Exec(ctx, q, usr.ID, usr.Name, usr.Email, usr.PasswordHash, usr.Roles, usr.DateCreated, usr.DateUpdated); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == uniqueViolation {
				// violates unique constraint "users_email_key"
				return Info{}, data.ErrDuplicateEmail
			}
		}
		return Info{}, errors.Wrap(err, "inserting user")
	}

	return usr, nil
}

// Update replaces a user document in the database.
func (s Store) Update(ctx context.Context, traceID string, claims auth.Claims, userID string, uu UpdateUser, now time.Time) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.user.update")
	defer span.End()

	usr, err := s.QueryByID(ctx, traceID, claims, userID)
	if err != nil {
		return err
	}

	conn, err := s.db.Acquire(ctx)
	if err != nil {
		return errors.Wrap(err, "acquire db connection")
	}
	defer conn.Release()

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
		"name" = $2,
		"email" = $3,
		"roles" = $4,
		"password_hash" = $5,
		"date_updated" = $6
	WHERE
		user_id = $1`

	if _, err = conn.Exec(ctx, q, userID, usr.Name, usr.Email, usr.Roles, usr.PasswordHash, usr.DateUpdated); err != nil {
		return errors.Wrap(err, "updating user")
	}

	return nil
}

// Delete removes a user from the database.
func (s Store) Delete(ctx context.Context, traceID string, claims auth.Claims, userID string) error {
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

	conn, err := s.db.Acquire(ctx)
	if err != nil {
		return errors.Wrap(err, "acquire db connection")
	}
	defer conn.Release()

	const q = `
	DELETE FROM
		users
	WHERE
		user_id = $1`

	if _, err := conn.Exec(ctx, q, userID); err != nil {
		return errors.Wrapf(err, "deleting user %s", userID)
	}

	return nil
}

// Query retrieves a list of existing users from the database.
func (s Store) Query(ctx context.Context, traceID string, pageNumber int, rowsPerPage int) ([]Info, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.user.query")
	defer span.End()

	conn, err := s.db.Acquire(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "acquire db connection")
	}
	defer conn.Release()

	const q = `
	SELECT
		*
	FROM
		users
	ORDER BY
		user_id
		OFFSET $1 ROWS FETCH NEXT $2 ROWS ONLY`

	page := struct {
		Offset      int `db:"offset"`
		RowsPerPage int `db:"rows_per_page"`
	}{
		Offset:      (pageNumber - 1) * rowsPerPage,
		RowsPerPage: rowsPerPage,
	}

	users := make([]Info, 0, page.RowsPerPage)
	if err := pgxscan.Select(ctx, conn, &users, q, page.Offset, page.RowsPerPage); err != nil {
		return nil, errors.Wrap(err, "query users")
	}

	return users, nil
}

// QueryByID gets the specified user from the database.
func (s Store) QueryByID(ctx context.Context, traceID string, claims auth.Claims, userID string) (Info, error) {
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

	conn, err := s.db.Acquire(ctx)
	if err != nil {
		return Info{}, errors.Wrap(err, "acquire db connection")
	}
	defer conn.Release()

	const q = `SELECT * FROM users WHERE user_id = $1`

	var usr Info
	if err := pgxscan.Get(ctx, conn, &usr, q, userID); err != nil {
		if pgxscan.NotFound(err) {
			return Info{}, data.ErrNotFound
		}
		return Info{}, errors.Wrapf(err, "selecting user %q", userID)
	}

	return usr, nil
}

// QueryByEmail gets the specified user from the database by email.
func (s Store) QueryByEmail(ctx context.Context, traceID string, claims auth.Claims, email string) (Info, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.user.querybyemail")
	defer span.End()

	conn, err := s.db.Acquire(ctx)
	if err != nil {
		return Info{}, errors.Wrap(err, "acquire db connection")
	}
	defer conn.Release()

	const q = `SELECT *	FROM users WHERE email = $1`

	var usr Info
	if err := pgxscan.Get(ctx, conn, &usr, q, email); err != nil {
		if pgxscan.NotFound(err) {
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
func (s Store) Authenticate(ctx context.Context, traceID string, now time.Time, email, password string) (auth.Claims, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.user.authenticate")
	defer span.End()

	conn, err := s.db.Acquire(ctx)
	if err != nil {
		return auth.Claims{}, errors.Wrap(err, "acquire db connection")
	}
	defer conn.Release()

	const q = `SELECT * FROM users WHERE email = $1`

	var usr Info
	if err := pgxscan.Get(ctx, conn, &usr, q, email); err != nil {
		// Normally we would return ErrNotFound in this scenario but we do not want
		// to leak to an unauthenticated user which emails are in the system.
		if pgxscan.NotFound(err) {
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
