package resolver

import "errors"

var (
	ErrNotFound  = errors.New("resolver: not found")
	ErrBadPath   = errors.New("resolver: bad path")
	ErrForbidden = errors.New("resolver: forbidden")
)
