package main

import (
	"context"
	"log"
	"net/http"

	"github.com/mcorrigan89/url_shortener/internal/usercontext"
	"github.com/rs/xid"
)

type contextKey string

const (
	ipKey            contextKey = "ip"
	correlationIDKey contextKey = "correlation_id"
	sessionTokenKey  contextKey = "session_token_key"
)

func (app *application) contextBuilder(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		ctx = context.WithValue(ctx, ipKey, r.RemoteAddr)
		correlationID := xid.New().String()
		ctx = context.WithValue(ctx, correlationIDKey, correlationID)

		ctx = app.logger.WithContext(ctx)

		sessionToken := r.Header.Get("x-session-token")
		ctx = context.WithValue(ctx, sessionTokenKey, sessionToken)

		ctx = app.logger.WithContext(ctx)

		user, session, err := app.services.UserService.GetUserBySessionToken(ctx, sessionToken)
		ctx = usercontext.ContextSetSession(ctx, session)
		if err == nil && !session.IsExpired() {
			ctx = usercontext.ContextSetUser(ctx, user)
		}

		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {

				w.Header().Set("Connect", "close")

				log.Fatalf("Panic not recovered %v \n", err)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *application) enabledCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")

		origin := r.Header.Get("Origin")

		if origin != "" {
			for i := range app.config.Cors.TrustedOrigins {
				if origin == app.config.Cors.TrustedOrigins[i] {

					w.Header().Set("Access-Control-Allow-Origin", origin)

					if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
						w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
						w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

						w.WriteHeader(http.StatusOK)
						return
					}
					break
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}
