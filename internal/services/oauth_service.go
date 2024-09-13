package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/mcorrigan89/url_shortener/internal/entities"
	"github.com/mcorrigan89/url_shortener/internal/repositories"
)

type OAuthService struct {
	utils          ServicesUtils
	userService    *UserService
	userRepository *repositories.UserRepository
}

func NewOAuthService(utils ServicesUtils, userService *UserService, userRepo *repositories.UserRepository) *OAuthService {
	return &OAuthService{
		utils:          utils,
		userService:    userService,
		userRepository: userRepo,
	}
}

type GoogleTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	IDToken      string `json:"id_token"`
}

func (service *OAuthService) LoginWithGoogleCode(ctx context.Context, code string) (*entities.UserSession, error) {
	tokenResponse, err := service.getGoogleTokenFromCode(ctx, code)
	if err != nil {
		service.utils.logger.Err(err).Ctx(ctx).Msg("Failed to get Google Token")
		return nil, err
	}

	user, err := service.greateGoogleUser(ctx, tokenResponse)
	if err != nil {
		service.utils.logger.Err(err).Ctx(ctx).Msg("Failed to create Google User")
		return nil, err
	}

	tokenJson, err := json.Marshal(tokenResponse)
	if err != nil {
		service.utils.logger.Err(err).Ctx(ctx).Msg("Failed to marshal Google Token")
		return nil, err
	}

	var userEntity *entities.User

	userEntity, err = service.userRepository.GetUserByProviderID(ctx, user.ID, "google")
	if err != nil {
		if err != entities.ErrUserNotFound {
			service.utils.logger.Err(err).Ctx(ctx).Msg("Failed to get User by Google ID")
			return nil, err
		} else {
			userEntity, err = service.userRepository.CreateUserOAuth(ctx, repositories.CreateUserOAuthArgs{
				GivenName:    user.GivenName,
				FamilyName:   user.FamilyName,
				Email:        user.Email,
				AvatarUrl:    user.Picture,
				Value:        tokenResponse.AccessToken,
				Provider:     "google",
				ProviderID:   user.ID,
				ProviderData: tokenJson,
			})

			if err != nil {
				service.utils.logger.Err(err).Ctx(ctx).Msg("Failed to create User from Google")
				return nil, err
			}
		}
	}

	userSessionEntity, err := service.userRepository.CreateUserSession(ctx, repositories.CreateUserSessionArgs{
		UserID: userEntity.ID,
	})

	if err != nil {
		service.utils.logger.Err(err).Ctx(ctx).Msg("Failed to create User Session from Google")
		return nil, err
	}

	return userSessionEntity, nil
}

func (service *OAuthService) getGoogleTokenFromCode(ctx context.Context, code string) (*GoogleTokenResponse, error) {
	service.utils.logger.Info().Ctx(ctx).Str("code", code).Msg("Getting Google Auth with Code")
	clientID := service.utils.config.OAuth.Google.ClientID
	clientSecret := service.utils.config.OAuth.Google.ClientSecret
	redirectUrl := fmt.Sprintf("%s/callback/google", service.utils.config.ClientURL)

	url := fmt.Sprintf("https://oauth2.googleapis.com/token?code=%s&client_id=%s&client_secret=%s&redirect_uri=%s&grant_type=authorization_code", code, clientID, clientSecret, redirectUrl)

	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		service.utils.logger.Err(err).Ctx(ctx).Msg("Failed to get Google Token")
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		service.utils.logger.Err(err).Ctx(ctx).Msg("Failed to read Google Token response")
		return nil, err
	}

	var tokenResponse GoogleTokenResponse
	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		service.utils.logger.Err(err).Ctx(ctx).Msg("Failed to unmarshal Google Token response")
		return nil, err
	}

	return &tokenResponse, nil
}

type googleUser struct {
	ID            string  `json:"id"`
	Email         string  `json:"email"`
	VerifiedEmail bool    `json:"verified_email"`
	Name          *string `json:"name,omitempty"`
	GivenName     *string `json:"given_name,omitempty"`
	FamilyName    *string `json:"family_name,omitempty"`
	Picture       *string `json:"picture,omitempty"`
}

func (service *OAuthService) greateGoogleUser(ctx context.Context, tokenResponse *GoogleTokenResponse) (*googleUser, error) {
	service.utils.logger.Info().Ctx(ctx).Msg("Creating Google User")

	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v1/userinfo", nil)
	if err != nil {
		service.utils.logger.Err(err).Ctx(ctx).Msg("Failed to create Google User request")
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenResponse.AccessToken))

	resp, err := client.Do(req)
	if err != nil {
		service.utils.logger.Err(err).Ctx(ctx).Msg("Failed to get Google User")
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		service.utils.logger.Err(err).Ctx(ctx).Msg("Failed to read Google User response")
		return nil, err
	}

	var user googleUser
	err = json.Unmarshal(body, &user)
	if err != nil {
		service.utils.logger.Err(err).Ctx(ctx).Msg("Failed to unmarshal Google User response")
		return nil, err
	}

	return &user, nil
}
