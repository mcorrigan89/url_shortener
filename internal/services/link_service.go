package services

import (
	"context"
	"errors"
	"math/rand"
	"net/url"

	"github.com/google/uuid"
	"github.com/mcorrigan89/url_shortener/internal/entities"
	"github.com/mcorrigan89/url_shortener/internal/repositories"
)

var (
	ErrBlockedDomain      = errors.New("domain is blocked")
	ErrBlockedEvent       = errors.New("event is blocked")
	ErrBlockedUser        = errors.New("user is blocked")
	ErrInvalidURL         = errors.New("invalid URL")
	ErrInvalidURLProtocol = errors.New("invalid URL protocol")
)

type LinkService struct {
	utils             ServicesUtils
	linkRepository    *repositories.LinkRepository
	blockedRepository *repositories.BlockedRepository
}

func NewLinkService(utils ServicesUtils, repos *repositories.Repositories) *LinkService {
	return &LinkService{
		utils:             utils,
		linkRepository:    repos.LinkRepository,
		blockedRepository: repos.BlockedRepository,
	}
}

func (service *LinkService) GetLinkByID(ctx context.Context, linkID uuid.UUID) (*entities.LinkEntity, error) {
	service.utils.logger.Info().Ctx(ctx).Str("linkID", linkID.String()).Msg("Getting link by ID")
	link, err := service.linkRepository.GetLinkByID(ctx, linkID)
	if err != nil {
		service.utils.logger.Err(err).Ctx(ctx).Msg("Error getting link by ID")
		return nil, err
	}

	return link, nil
}

func (service *LinkService) GetLinkByShortenedURL(ctx context.Context, shortenedURL string) (*entities.LinkEntity, error) {
	service.utils.logger.Info().Ctx(ctx).Str("shortenedURL", shortenedURL).Msg("Getting link by shortend URL")
	link, err := service.linkRepository.GetLinkByShortenedURL(ctx, shortenedURL)
	if err != nil {
		service.utils.logger.Err(err).Ctx(ctx).Msg("Error getting link by shortend URL")
		return nil, err
	}

	return link, nil
}

func (service *LinkService) GetLinksByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.LinkEntity, error) {
	service.utils.logger.Info().Ctx(ctx).Str("userID", userID.String()).Msg("Getting link by user id")
	links, err := service.linkRepository.GetLinksByUserID(ctx, userID)
	if err != nil {
		service.utils.logger.Err(err).Ctx(ctx).Msg("Error getting link by user id")
		return nil, err
	}

	return links, nil
}

type CreateLinkArgs struct {
	UserID  uuid.UUID
	LinkURL string
}

func (service *LinkService) CreateLink(ctx context.Context, args CreateLinkArgs) (*entities.LinkEntity, error) {
	service.utils.logger.Info().Ctx(ctx).Interface("args", args).Msg("Creating link")

	url, err := url.Parse(args.LinkURL)
	if err != nil {
		service.utils.logger.Err(err).Ctx(ctx).Msg("Error parsing URL")
		return nil, err
	}

	if url.Host == "" {
		service.utils.logger.Err(ErrInvalidURL).Ctx(ctx).Str("linkURL", args.LinkURL).Msg("Invalid URL")
		return nil, ErrInvalidURL
	}

	if url.Scheme != "https" {
		service.utils.logger.Err(ErrInvalidURLProtocol).Ctx(ctx).Str("linkURL", args.LinkURL).Msg("Invalid URL protocol")
		return nil, ErrInvalidURL
	}

	err = service.checkIfBlocked(ctx, checkBlockedArgs{
		Domain: url.Host,
		UserID: args.UserID,
	})
	if err != nil {
		return nil, err
	}

	shortendUrlSlug := service.generateShortenedURLSlug()
	link, err := service.linkRepository.CreateLink(ctx, repositories.CreateLinkArgs{
		LinkURL:    args.LinkURL,
		ShortedURL: shortendUrlSlug,
		CreatedBy:  args.UserID,
	})
	if err != nil {
		service.utils.logger.Err(err).Ctx(ctx).Msg("Error creating link")
		return nil, err
	}

	return link, nil
}

func (service *LinkService) generateShortenedURLSlug() string {
	var randomChars = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0987654321")

	b := make([]rune, 12)
	for i := range b {
		b[i] = randomChars[rand.Intn(len(randomChars))]
	}
	return string(b)
}

type UpdateLinkArgs struct {
	UserID  uuid.UUID
	LinkID  uuid.UUID
	LinkURL *string
	Active  *bool
}

func (service *LinkService) UpdateLink(ctx context.Context, args UpdateLinkArgs) (*entities.LinkEntity, error) {
	service.utils.logger.Info().Ctx(ctx).Interface("args", args).Msg("Updating link")

	link, err := service.linkRepository.UpdateLink(ctx, repositories.UpdateLinkArgs{
		ID:        args.LinkID,
		LinkURL:   args.LinkURL,
		Active:    args.Active,
		UpdatedBy: args.UserID,
	})
	if err != nil {
		service.utils.logger.Err(err).Ctx(ctx).Msg("Error creating link")
		return nil, err
	}

	return link, nil
}

func (service *LinkService) IsDomainBlocked(ctx context.Context, linkUrl string) error {

	url, err := url.Parse(linkUrl)
	if err != nil {
		service.utils.logger.Err(err).Ctx(ctx).Msg("Error parsing URL")
		return err
	}

	blocked, err := service.blockedRepository.IsDomainBlocked(ctx, url.Host)
	if err != nil {
		return err
	}
	if blocked {
		service.utils.logger.Err(ErrBlockedDomain).Ctx(ctx).Str("domain", url.Host).Str("linkUrl", linkUrl).Msg("Attmept to query a link with a blocked domain")
		return ErrBlockedDomain
	}

	return nil
}

type checkBlockedArgs struct {
	Domain  string
	UserID  uuid.UUID
	EventID *uuid.UUID
}

func (service *LinkService) checkIfBlocked(ctx context.Context, args checkBlockedArgs) error {
	service.utils.logger.Info().Ctx(ctx).Interface("args", args).Msg("Checking if blocked")
	blocked, err := service.blockedRepository.IsDomainBlocked(ctx, args.Domain)
	if err != nil {
		return err
	}
	if blocked {
		service.utils.logger.Err(ErrBlockedDomain).Ctx(ctx).Interface("args", args).Msg("Attmept to create link with blocked domain")
		return ErrBlockedDomain
	}

	blocked, err = service.blockedRepository.IsUserBlocked(ctx, args.UserID)
	if err != nil {
		return err
	}
	if blocked {
		service.utils.logger.Err(ErrBlockedUser).Ctx(ctx).Interface("args", args).Msg("Attmept to create link with blocked user")
		return ErrBlockedUser
	}

	return nil
}
