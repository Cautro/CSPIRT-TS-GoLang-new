package postgres

import "errors"

var (
	ErrBadRequest = errors.New("Bad request")
	ErrServer = errors.New("Server error")
)