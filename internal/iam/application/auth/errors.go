package auth

import "errors"

var (
	ErrPwdWrong        = errors.New("error when logging in")
	ErrTokenDuplicated = errors.New("error when logging in")
)
