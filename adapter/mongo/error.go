package mongo

import "errors"

var (
	// ErrNotFound is returned when the requested resource is not found.
	ErrNotFound = errors.New("not found")
	// ErrBadRequest is returned when the request is malformed.
	ErrBadRequest = errors.New("bad request")
	// ErrUnprocessableEntity is returned when the request is formed correctly but not valid.
	ErrUnprocessableEntity = errors.New("unprocessable entity")
)
