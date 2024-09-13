package entities

import (
	"time"

	"github.com/google/uuid"
)

type UserSession struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	Token         string
	ExpiresAt     time.Time
	ExpiredByUser bool
}

type NewUserSessionArgs struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	Token         string
	ExpiresAt     time.Time
	ExpiredByUser bool
}

func NewUserSession(args NewUserSessionArgs) *UserSession {
	return &UserSession{
		ID:            args.ID,
		UserID:        args.UserID,
		Token:         args.Token,
		ExpiresAt:     args.ExpiresAt,
		ExpiredByUser: args.ExpiredByUser,
	}
}

func (t *UserSession) IsExpired() bool {
	if t.ExpiredByUser {
		return true
	}
	return t.ExpiresAt.Before(time.Now())
}
