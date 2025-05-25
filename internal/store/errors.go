package store

import "errors"

var (
	ErrHumanNotFound   = errors.New("human not found")
	ErrNothingToUpdate = errors.New("nothing to update")
)
