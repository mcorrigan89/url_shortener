package services

import (
	"sync"

	"github.com/mcorrigan89/url_shortener/internal/config"
	"github.com/mcorrigan89/url_shortener/internal/repositories"
	"github.com/rs/zerolog"
)

type ServicesUtils struct {
	logger *zerolog.Logger
	wg     *sync.WaitGroup
	config *config.Config
}

type Services struct {
	utils        ServicesUtils
	UserService  *UserService
	OAuthService *OAuthService
	LinkService  *LinkService
}

func (utils *ServicesUtils) background(fn func()) {
	utils.wg.Add(1)

	go func() {
		defer utils.wg.Done()

		defer func() {
			if err := recover(); err != nil {
				utils.logger.Error().Msg("panic in background function")
			}
		}()

		fn()
	}()
}

func NewServices(repositories *repositories.Repositories, cfg *config.Config, logger *zerolog.Logger, wg *sync.WaitGroup) Services {
	utils := ServicesUtils{
		logger: logger,
		wg:     wg,
		config: cfg,
	}

	userService := NewUserService(utils, repositories.UserRepository)
	oAuthService := NewOAuthService(utils, userService, repositories.UserRepository)
	linkService := NewLinkService(utils, repositories)

	return Services{
		utils:        utils,
		UserService:  userService,
		OAuthService: oAuthService,
		LinkService:  linkService,
	}
}
