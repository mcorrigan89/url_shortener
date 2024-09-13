package dto

import "github.com/mcorrigan89/url_shortener/internal/validator"

type CreateLinkForm struct {
	LinkUrl             string `form:"link_url"`
	validator.Validator `form:"-"`
}
