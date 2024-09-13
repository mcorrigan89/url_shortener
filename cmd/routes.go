package main

import (
	"net/http"
)

func (app *application) routes() http.Handler {

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthy", app.healthy)
	mux.HandleFunc("GET /ping", app.ping)
	mux.HandleFunc("POST /link", app.createLink)
	mux.HandleFunc("GET /go/{slug}", app.redirectHandler)
	mux.HandleFunc("GET /link/{slug}", app.redirectHandler)
	mux.HandleFunc("GET /qr/{id}", app.qrCodeHandler)

	return app.recoverPanic(app.enabledCORS(app.contextBuilder(mux)))
}
