package entities

import (
	"errors"

	"github.com/google/uuid"
	"github.com/mcorrigan89/url_shortener/internal/repositories/models"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrNoPasswordSet      = errors.New("no password set")
	ErrIncorrectProvider  = errors.New("incorrect provider")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrDuplicateEmail     = errors.New("duplicate email")
)

var (
	ProviderPassword = "password"
)

type UserAuth struct {
	Value    string
	Provider string
}

func (ua *UserAuth) CompareHashAndPassword(password string) error {
	if ua.Provider != "password" {
		return ErrIncorrectProvider
	}
	err := bcrypt.CompareHashAndPassword([]byte(ua.Value), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidCredentials
		} else {
			return err
		}
	}

	return nil
}

type User struct {
	ID         uuid.UUID
	GivenName  *string
	FamilyName *string
	Email      string
	AvatarUrl  *string
	userAuth   *UserAuth
}

type NewUserEntityArgs struct {
	ID         uuid.UUID
	GivenName  *string
	FamilyName *string
	Email      string
	AvatarUrl  *string
	UserAuth   *UserAuth
}

func NewUserEntityFromModel(userModel models.User, userAuthModel models.UserAuth) *User {
	entity := &User{
		ID:         userModel.ID,
		GivenName:  userModel.GivenName,
		FamilyName: userModel.FamilyName,
		Email:      userModel.Email,
		AvatarUrl:  userModel.AvatarUrl,
		userAuth: &UserAuth{
			Value:    userAuthModel.Value,
			Provider: userAuthModel.Provider,
		},
	}

	return entity
}

func (u *User) ComparePassword(password string) error {
	return u.userAuth.CompareHashAndPassword(password)
}
