package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/mcorrigan89/url_shortener/internal/entities"
	"github.com/mcorrigan89/url_shortener/internal/repositories"
	"github.com/mcorrigan89/url_shortener/internal/usercontext"
)

type UserService struct {
	utils          ServicesUtils
	userRepository *repositories.UserRepository
}

func NewUserService(utils ServicesUtils, userRepo *repositories.UserRepository) *UserService {
	return &UserService{
		utils:          utils,
		userRepository: userRepo,
	}
}

func (service *UserService) GetUserByID(ctx context.Context, userId uuid.UUID) (*entities.User, error) {
	service.utils.logger.Info().Ctx(ctx).Str("userId", userId.String()).Msg("Getting user by ID")

	user, err := service.userRepository.GetUserByID(ctx, userId)
	if err != nil {
		service.utils.logger.Err(err).Ctx(ctx).Msg("Failed to get user by ID")
		return nil, err
	}

	return user, nil
}

func (service *UserService) GetUserBySessionToken(ctx context.Context, token string) (*entities.User, *entities.UserSession, error) {
	service.utils.logger.Info().Ctx(ctx).Str("token", token).Msg("Getting user by token")
	user, session, err := service.userRepository.GetUserBySessionToken(ctx, token)
	if err != nil {
		if err == entities.ErrUserNotFound {
			return nil, nil, err
		} else {
			service.utils.logger.Err(err).Ctx(ctx).Msg("Failed to get user by token")
			return nil, nil, err
		}
	}

	return user, session, nil
}

type CreateUserArgs struct {
	GivenName  *string
	FamilyName *string
	Email      string
	Password   string
}

func (service *UserService) CreateUser(ctx context.Context, args CreateUserArgs) (*entities.User, error) {
	service.utils.logger.Info().Ctx(ctx).Interface("args", args).Msg("Creating user")
	user, err := service.userRepository.CreateUserPassword(ctx, repositories.CreateUserPasswordArgs{
		GivenName:  args.GivenName,
		FamilyName: args.FamilyName,
		Email:      args.Email,
		Password:   args.Password,
	})
	if err != nil {
		service.utils.logger.Err(err).Ctx(ctx).Msg("Failed to create user")
		return nil, err
	}

	return user, nil
}

func (service *UserService) AuthenticateWithPassword(ctx context.Context, email string, password string) (*entities.UserSession, error) {
	service.utils.logger.Info().Ctx(ctx).Str("email", email).Msg("Authenticating user with password")
	user, err := service.userRepository.GetUserByEmail(ctx, email)
	if err != nil {
		service.utils.logger.Err(err).Ctx(ctx).Msg("Failed to get user by email")
		return nil, err
	}

	err = user.ComparePassword(password)
	if err != nil {
		service.utils.logger.Err(err).Ctx(ctx).Msg("Failed to compare password")
		return nil, err
	}

	session, err := service.userRepository.CreateUserSession(ctx, repositories.CreateUserSessionArgs{
		UserID: user.ID,
	})
	if err != nil {
		service.utils.logger.Err(err).Ctx(ctx).Msg("Failed to create user session")
		return nil, err
	}

	currentSession := usercontext.ContextGetSession(ctx)

	if currentSession != nil {
		service.userRepository.ExpireUserSession(ctx, currentSession.ID)
	}

	return session, nil
}
