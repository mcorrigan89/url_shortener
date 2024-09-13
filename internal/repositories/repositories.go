package repositories

import (
	"errors"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mcorrigan89/url_shortener/internal/config"
	"github.com/mcorrigan89/url_shortener/internal/repositories/models"
	"github.com/rs/zerolog"
)

const defaultTimeout = 10 * time.Second

var (
	ErrNotFound = errors.New("not found")
)

type ServicesUtils struct {
	logger *zerolog.Logger
	wg     *sync.WaitGroup
	cfg    *config.Config
}

type Repositories struct {
	utils             ServicesUtils
	UserRepository    *UserRepository
	LinkRepository    *LinkRepository
	BlockedRepository *BlockedRepository
}

func NewRepositories(db *pgxpool.Pool, cfg *config.Config, logger *zerolog.Logger, wg *sync.WaitGroup) Repositories {
	queries := models.New(db)
	utils := ServicesUtils{
		logger: logger,
		wg:     wg,
		cfg:    cfg,
	}

	userRepo := NewUserRepository(utils, db, queries)
	linkRepo := NewLinkRepository(utils, db, queries)
	blockRepo := NewBlockedRepository(utils, db, queries)

	return Repositories{
		utils:             utils,
		UserRepository:    userRepo,
		LinkRepository:    linkRepo,
		BlockedRepository: blockRepo,
	}
}
