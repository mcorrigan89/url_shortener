package main

import (
	"context"
	"os"
	"time"

	"github.com/mcorrigan89/url_shortener/internal/config"

	"github.com/rs/zerolog"
)

func getCorrelationIdFromContext(ctx context.Context) string {
	correlationId, ok := ctx.Value(correlationIDKey).(string)
	if !ok {
		return ""
	}
	return correlationId
}

func getSessionTokenFromContext(ctx context.Context) string {
	sessionToken, ok := ctx.Value(sessionTokenKey).(string)
	if !ok {
		return ""
	}
	return sessionToken
}

func getIPFromContext(ctx context.Context) string {
	ip, ok := ctx.Value(ipKey).(string)
	if !ok {
		return ""
	}
	return ip
}

type TracingHook struct{}

func (h TracingHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	ctx := e.GetCtx()
	correlationId := getCorrelationIdFromContext(ctx)
	if correlationId != "" {
		e.Str("correlation_id", correlationId)
	}
	ip := getIPFromContext(ctx)
	if ip != "" {
		e.Str("ip_address", ip)
	}
}

func getLogger(cfg config.Config) zerolog.Logger {

	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).
		With().
		Timestamp().
		Caller().
		Stack().
		Logger().
		Hook(TracingHook{})

	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	return logger

}
