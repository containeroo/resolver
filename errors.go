package resolver

import (
	"errors"
	"fmt"
	"io/fs"
)

var (
	ErrNotFound   = errors.New("resolver: not found")
	ErrBadPath    = errors.New("resolver: bad path")
	ErrForbidden  = errors.New("resolver: forbidden")
	ErrEmptyValue = errors.New("resolver: empty value")
)

func classifyFileError(err error) error {
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return fmt.Errorf("%w: %w", ErrNotFound, err)
	case errors.Is(err, fs.ErrPermission):
		return fmt.Errorf("%w: %w", ErrForbidden, err)
	default:
		return err
	}
}
