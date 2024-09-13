package main

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/mcorrigan89/url_shortener/internal/services"
	"github.com/mcorrigan89/url_shortener/internal/usercontext"
	"github.com/skip2/go-qrcode"
)

func (app *application) ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func (app *application) healthy(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`{"healthy": true}`))
}

func (app *application) redirectHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	slugParam := r.PathValue("slug")

	linkEntity, err := app.services.LinkService.GetLinkByShortenedURL(ctx, slugParam)
	if err != nil {
		app.logger.Err(err).Ctx(ctx).Msg("Error getting link by shortened URL")
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if !linkEntity.Active {
		app.logger.Warn().Ctx(ctx).Msg("Deactived link requested")
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if linkEntity.Quarantined {
		app.logger.Warn().Ctx(ctx).Msg("Quarantined link requested")
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	err = app.services.LinkService.IsDomainBlocked(ctx, linkEntity.LinkURL)
	if err != nil {
		app.logger.Err(err).Ctx(ctx).Msg("Error checking if domain is blocked")
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	app.logger.Info().Ctx(ctx).Str("linkURL", linkEntity.LinkURL).Str("slug", slugParam).Msg("Link visited")

	http.Redirect(w, r, linkEntity.LinkURL, http.StatusFound)
}

func (app *application) qrCodeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	linkIDParam := r.PathValue("id")

	linkUUID, err := uuid.Parse(linkIDParam)
	if err != nil {
		app.logger.Err(err).Ctx(ctx).Msg("Error parsing link ID")
		http.Error(w, "Malformed UUID", http.StatusBadRequest)
		return
	}

	linkEntity, err := app.services.LinkService.GetLinkByID(ctx, linkUUID)
	if err != nil {
		app.logger.Err(err).Ctx(ctx).Msg("Error getting link by shortened URL")
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	redirectURL := fmt.Sprintf("%s/go/%s", app.config.ClientURL, linkEntity.ShortenedURLSlug)

	code, err := qrcode.Encode(redirectURL, qrcode.Medium, 256)
	if err != nil {
		app.logger.Err(err).Ctx(ctx).Msg("Error generating QR code")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Write(code)
}

func (app *application) createLink(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user := usercontext.ContextGetUser(ctx)

	if user == nil {
		app.logger.Warn().Ctx(ctx).Msg("Unauthenticated user")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var input struct {
		LinkURL string `json:"link_url"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.logger.Err(err).Ctx(ctx).Msg("Error reading JSON")
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	fmt.Println(input)

	_, err = app.services.LinkService.CreateLink(ctx, services.CreateLinkArgs{
		UserID:  user.ID,
		LinkURL: input.LinkURL,
	})

	if err != nil {
		app.logger.Err(err).Ctx(ctx).Msg("Error creating link")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}
