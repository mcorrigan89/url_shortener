package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mcorrigan89/url_shortener/internal/entities"
	"github.com/mcorrigan89/url_shortener/internal/repositories/models"
)

type LinkRepository struct {
	utils   ServicesUtils
	DB      *pgxpool.Pool
	queries *models.Queries
}

func NewLinkRepository(utils ServicesUtils, db *pgxpool.Pool, queries *models.Queries) *LinkRepository {
	return &LinkRepository{
		utils:   utils,
		DB:      db,
		queries: queries,
	}
}

func (repo *LinkRepository) GetLinkByID(ctx context.Context, linkID uuid.UUID) (*entities.LinkEntity, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	linkRow, err := repo.queries.GetLinkByID(ctx, linkID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		} else {
			repo.utils.logger.Err(err).Ctx(ctx).Msg("Error getting link by ID")
			return nil, err
		}
	}

	link := repo.modelToEntity(linkRow)

	return &link, nil
}

func (repo *LinkRepository) GetLinkByShortenedURL(ctx context.Context, shortenedUrlSlug string) (*entities.LinkEntity, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	linkRow, err := repo.queries.GetLinkByShortenedURL(ctx, shortenedUrlSlug)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		} else {
			repo.utils.logger.Err(err).Ctx(ctx).Msg("Error getting link by shortend URL")
			return nil, err
		}
	}

	link := repo.modelToEntity(linkRow)

	return &link, nil
}

func (repo *LinkRepository) GetLinksByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.LinkEntity, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	linkRows, err := repo.queries.GetLinksByUserID(ctx, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		} else {
			repo.utils.logger.Err(err).Ctx(ctx).Str("userID", userID.String()).Msg("Error getting link by user id")
			return nil, err
		}
	}

	links := []*entities.LinkEntity{}

	for _, linkRow := range linkRows {
		link := repo.modelToEntity(linkRow)
		links = append(links, &link)
	}

	return links, nil
}

type CreateLinkArgs struct {
	LinkURL    string
	ShortedURL string
	CreatedBy  uuid.UUID
}

func (repo *LinkRepository) CreateLink(ctx context.Context, args CreateLinkArgs) (*entities.LinkEntity, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	linkRow, err := repo.queries.CreateLink(ctx, models.CreateLinkParams{
		LinkUrl:      args.LinkURL,
		CreatedBy:    args.CreatedBy,
		UpdatedBy:    args.CreatedBy,
		ShortenedUrl: args.ShortedURL,
	})

	if err != nil {
		repo.utils.logger.Err(err).Ctx(ctx).Msg("Error creating link")
		return nil, err
	}

	link := repo.modelToEntity(linkRow)

	return &link, nil
}

type UpdateLinkArgs struct {
	ID        uuid.UUID
	UpdatedBy uuid.UUID
	LinkURL   *string
	Active    *bool
}

func (repo *LinkRepository) UpdateLink(ctx context.Context, args UpdateLinkArgs) (*entities.LinkEntity, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	tx, err := repo.DB.Begin(ctx)
	if err != nil {
		repo.utils.logger.Err(err).Ctx(ctx).Msg("Error with transaction updating link")
		return nil, err
	}
	defer tx.Rollback(ctx)

	qtx := repo.queries.WithTx(tx)

	previousLinkRow, err := qtx.GetLinkByID(ctx, args.ID)
	if err != nil {
		repo.utils.logger.Err(err).Ctx(ctx).Msg("Error getting previous link")
		return nil, err
	}

	updatedLinkRow, err := qtx.UpdateLink(ctx, models.UpdateLinkParams{
		ID:        args.ID,
		LinkUrl:   args.LinkURL,
		UpdatedBy: args.UpdatedBy,
		Active:    args.Active,
	})
	if err != nil {
		repo.utils.logger.Err(err).Ctx(ctx).Msg("Error updating link")
		return nil, err
	}

	_, err = qtx.CreateLinkHistory(ctx, models.CreateLinkHistoryParams{
		LinkID:      previousLinkRow.ID,
		LinkUrl:     previousLinkRow.LinkUrl,
		Active:      previousLinkRow.Active,
		Quarantined: previousLinkRow.Quarantined,
		CreatedBy:   previousLinkRow.CreatedBy,
		UpdatedBy:   args.UpdatedBy,
	})
	if err != nil {
		repo.utils.logger.Err(err).Ctx(ctx).Msg("Error creating link history")
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		repo.utils.logger.Err(err).Ctx(ctx).Msg("Error committing transaction")
		return nil, err
	}

	link := repo.modelToEntity(updatedLinkRow)

	return &link, nil
}

func (repo *LinkRepository) modelToEntity(model models.LinkRedirect) entities.LinkEntity {
	return entities.LinkEntity{
		ID:               model.ID,
		ShortenedURL:     fmt.Sprintf("%s/go/%s", repo.utils.cfg.ClientURL, model.ShortenedUrl),
		ShortenedURLSlug: model.ShortenedUrl,
		LinkURL:          model.LinkUrl,
		CreatedBy:        model.CreatedBy,
		Quarantined:      model.Quarantined,
		Active:           model.Active,
	}
}
