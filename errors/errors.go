package errors

import "errors"

var (
	ErrNotFound          = errors.New("not found")
	ErrNilArgument       = errors.New("nil argument")
	ErrFailedToDecodeHex = errors.New("failed to decode hex")
)

func Is(err error, target error) bool {
	return errors.Is(err, target)
}
