package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mcorrigan89/url_shortener/internal/repositories/models"
)

type BlockedRepository struct {
	utils   ServicesUtils
	DB      *pgxpool.Pool
	queries *models.Queries
}

func NewBlockedRepository(utils ServicesUtils, db *pgxpool.Pool, queries *models.Queries) *BlockedRepository {
	return &BlockedRepository{
		utils:   utils,
		DB:      db,
		queries: queries,
	}
}

func (repo *BlockedRepository) IsDomainBlocked(ctx context.Context, domain string) (bool, error) {
	_, err := repo.queries.GetBlockedDomain(ctx, domain)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		} else {
			repo.utils.logger.Err(err).Ctx(ctx).Msg("Error checking if domain is blocked")
			return true, err
		}
	}
	return true, nil
}

func (repo *BlockedRepository) IsUserBlocked(ctx context.Context, userID uuid.UUID) (bool, error) {
	_, err := repo.queries.GetBlockedUser(ctx, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		} else {
			repo.utils.logger.Err(err).Ctx(ctx).Msg("Error checking if user is blocked")
			return true, err
		}
	}
	return true, nil
}
