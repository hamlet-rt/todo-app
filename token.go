package todo

import "time"

type RefreshToken struct {
	ID        int64
	UserID    int64
	Token     string
	ExpiresDate time.Time
}