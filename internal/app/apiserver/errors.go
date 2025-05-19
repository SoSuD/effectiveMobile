package apiserver

import "errors"

var (
	ErrNameAndSurnameRequired = errors.New("name and surname required")
)
