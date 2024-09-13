package entities

import "github.com/google/uuid"

type LinkEntity struct {
	ID               uuid.UUID
	ShortenedURL     string
	ShortenedURLSlug string
	LinkURL          string
	CreatedBy        uuid.UUID
	Active           bool
	Quarantined      bool
}
