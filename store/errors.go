package store

import "errors"

var (
	// ErrTooManyAffectedRows returned when store modifes too many rows.
	ErrTooManyAffectedRows = errors.New("too many affected rows")
)
