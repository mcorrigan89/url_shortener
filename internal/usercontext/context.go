package usercontext

import (
	"context"

	"github.com/google/uuid"
	"github.com/mcorrigan89/url_shortener/internal/entities"
)

var AnonymousUser = &entities.User{
	ID: uuid.Nil,
}

func UserIsAnonymous(u entities.User) bool {
	return AnonymousUser.ID == u.ID
}

type contextKey string

const currentUserContextKey = contextKey("currentUser")
const currentSessionContextKey = contextKey("currentSession")

func ContextSetUser(ctx context.Context, user *entities.User) context.Context {
	ctx = context.WithValue(ctx, currentUserContextKey, user)
	return ctx
}

func ContextGetUser(ctx context.Context) *entities.User {
	user, ok := ctx.Value(currentUserContextKey).(*entities.User)
	if !ok {
		return nil
	}

	return user
}

func ContextSetSession(ctx context.Context, session *entities.UserSession) context.Context {
	ctx = context.WithValue(ctx, currentSessionContextKey, session)
	return ctx
}

func ContextGetSession(ctx context.Context) *entities.UserSession {
	session, ok := ctx.Value(currentSessionContextKey).(*entities.UserSession)
	if !ok {
		return nil
	}

	return session
}
