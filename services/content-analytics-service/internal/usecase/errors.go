package usecase

import "errors"

type InvalidInputError struct{ msg string }

func (e InvalidInputError) Error() string { return e.msg }

func ErrInvalidInputError(msg string) error {
	return InvalidInputError{msg: msg}
}

func IsInvalidInput(err error) bool {
	var e InvalidInputError
	return errors.As(err, &e)
}