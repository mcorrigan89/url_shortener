package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func (app *application) serve() error {

	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	routes := h2c.NewHandler(app.routes(), &http2.Server{})

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.Port),
		Handler:      routes,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		TLSConfig:    tlsConfig,
	}

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)

		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		app.logger.Info().Str("signal", s.String()).Msg("caught signal")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			app.logger.Err(err).Msg("graceful shutdown failed")
			shutdownError <- err
		}

		app.logger.Info().Str("addr", srv.Addr).Msg("completing background tasks")

		app.wg.Wait()
		shutdownError <- nil
	}()

	app.logger.Info().Str("addr", srv.Addr).Str("env", app.config.Env).Msg("listening on")

	err := srv.ListenAndServe()

	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	app.logger.Info().Str("addr", srv.Addr).Msg("stopped server")

	return nil
}
