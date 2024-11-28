package todo

import "errors"

var (
	ErrRefreshTokenExpired = errors.New("refresh token expired")
)