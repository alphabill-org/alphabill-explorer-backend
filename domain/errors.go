package domain

import "errors"

var (
	ErrNotFound          = errors.New("not found")
	ErrNilArgument       = errors.New("nil argument")
	ErrFailedToDecodeHex = errors.New("failed to decode hex")
)
