package main

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/mcorrigan89/url_shortener/dto"
	"github.com/mcorrigan89/url_shortener/internal/services"
	"github.com/mcorrigan89/url_shortener/internal/usercontext"
	"github.com/mcorrigan89/url_shortener/internal/validator"
	"github.com/mcorrigan89/url_shortener/ui"
	"github.com/skip2/go-qrcode"
)

func (app *application) ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func (app *application) healthy(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`{"healthy": true}`))
}

func (app *application) homePage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	home := ui.Base("Home", "Home page", ui.Home())

	home.Render(ctx, w)
}

func (app *application) loginPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	login := ui.Base("Login", "Login page", ui.Login(app.config))

	login.Render(ctx, w)
}

func (app *application) linksPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user := usercontext.ContextGetUser(ctx)

	if user == nil {
		app.logger.Warn().Ctx(ctx).Msg("Unauthenticated user")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	linkEntities, err := app.services.LinkService.GetLinksByUserID(ctx, user.ID)
	if err != nil {
		app.logger.Err(err).Ctx(ctx).Msg("Error getting links by user ID")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	links := ui.Base("My Links", "Links page", ui.Links(app.config, linkEntities))

	links.Render(ctx, w)
}

func (app *application) createLinkPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	createLink := ui.Base("Create link", "Create link page", ui.CreateLink(dto.CreateLinkForm{}))

	createLink.Render(ctx, w)
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

func (app *application) loginPassword(w http.ResponseWriter, r *http.Request) {
	// ctx := r.Context()

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message": "You are authenticated"}`))
}

func (app *application) loginGoogle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	queryParams := r.URL.Query()

	codeParam := queryParams.Get("code")

	if codeParam == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	session, err := app.services.OAuthService.LoginWithGoogleCode(ctx, codeParam)

	if err != nil {
		app.logger.Err(err).Ctx(ctx).Msg("Error logging in with Google")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	cookie := http.Cookie{
		Name:     "x-session-token",
		Value:    session.Token,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
	}
	http.SetCookie(w, &cookie)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type createLinkForm struct {
	LinkURL string `form:"link_url"`
}

func (app *application) createLink(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user := usercontext.ContextGetUser(ctx)

	if user == nil {
		app.logger.Warn().Ctx(ctx).Msg("Unauthenticated user")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var form dto.CreateLinkForm

	err := r.ParseForm()
	if err != nil {
		app.logger.Err(err).Ctx(ctx).Msg("Error parsing form")
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	err = app.formDecoder.Decode(&form, r.PostForm)

	if err != nil {
		app.logger.Err(err).Ctx(ctx).Msg("Error decoding form")
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.LinkUrl), "link_url", "This field cannot be blank")
	form.CheckField(validator.IsValidHTTPSDomainAndURL(form.LinkUrl), "link_url", "This field must be a valid URL")

	if !form.Valid() {
		fmt.Println(form.FieldErrors["link_url"])
		createLink := ui.Base("Create link", "Create link page", ui.CreateLink(form))

		createLink.Render(ctx, w)
		return
	}

	_, err = app.services.LinkService.CreateLink(ctx, services.CreateLinkArgs{
		UserID:  user.ID,
		LinkURL: form.LinkUrl,
	})

	if err != nil {
		app.logger.Err(err).Ctx(ctx).Msg("Error creating link")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/links", http.StatusSeeOther)
}
