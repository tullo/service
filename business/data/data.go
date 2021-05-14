package data

import "errors"

// Set of error variables.
var (
	ErrNotFound = errors.New("not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("ID is not in its proper form")

	// ErrForbidden occurs when a user tries to do something
	// that is forbidden to them according to our access
	// control policies.
	ErrForbidden = errors.New("attempted action is not allowed")

	// ErrAuthenticationFailure occurs when a user attempts
	// to authenticate but anything goes wrong.
	ErrAuthenticationFailure = errors.New("authentication failed")

	// ErrDuplicateEmail occurs when user creation failed
	// b/c of an email address that's already in use.
	ErrDuplicateEmail = errors.New("duplicate email")
)
