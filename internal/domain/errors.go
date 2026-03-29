package domain

import "errors"

var (
	ErrProductVersionConflict = errors.New("product version conflict: document was modified by another request")
	ErrOrderVersionConflict   = errors.New("order version conflict: document was modified by another request")
)
