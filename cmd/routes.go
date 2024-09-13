package main

import (
	"net/http"

	"github.com/mcorrigan89/url_shortener/public"
)

func (app *application) routes() http.Handler {

	mux := http.NewServeMux()

	fs := public.Files
	fileServer := http.FileServer(http.FS(fs))
	mux.Handle("/static/*", fileServer)
	mux.Handle("/favicon.ico", fileServer)
	mux.Handle("/robots.txt", fileServer)
	mux.HandleFunc("GET /healthy", app.healthy)
	mux.HandleFunc("GET /ping", app.ping)

	// Pages
	mux.HandleFunc("/", app.homePage)
	mux.HandleFunc("/login", app.loginPage)
	mux.HandleFunc("/links", app.linksPage)
	mux.HandleFunc("/create", app.createLinkPage)

	// Operations
	mux.HandleFunc("GET /callback/google", app.loginGoogle)
	mux.HandleFunc("POST /login/password", app.loginPassword)
	mux.HandleFunc("POST /create", app.createLink)

	// Redirects
	mux.HandleFunc("GET /go/{slug}", app.redirectHandler)
	mux.HandleFunc("GET /link/{slug}", app.redirectHandler)
	mux.HandleFunc("GET /qr/{id}", app.qrCodeHandler)

	return app.recoverPanic(app.enabledCORS(app.contextBuilder(mux)))
}
