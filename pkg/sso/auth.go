package sso

import (
	"context"
	"errors"
	"google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
	"net/http"
)

type GoogleTokenInterface interface {
	Validation(context.Context, string) (*oauth2.Tokeninfo, error)
}

type GoogleToken struct {
	googleClientId string
}

func NewGoogleToken(googleClientId string) *GoogleToken {
	return &GoogleToken{googleClientId}
}

func (g *GoogleToken) Validation(ctx context.Context, token string) (*oauth2.Tokeninfo, error) {
	oauth2Service, err := oauth2.NewService(ctx, option.WithHTTPClient(http.DefaultClient))
	if err != nil {
		return nil, errors.New("error to create oauth service")
	}

	tokenInfo, err := oauth2Service.Tokeninfo().IdToken(token).Do()
	if err != nil {
		return nil, errors.New("invalid token info")
	}

	if g.googleClientId != tokenInfo.Audience {
		return nil, errors.New("invalid client")
	}

	return tokenInfo, nil
}

func ValidateGoogleToken(token string) (*oauth2.Tokeninfo, error) {

	oauth2Service, err := oauth2.NewService(context.Background(), option.WithHTTPClient(http.DefaultClient))
	if err != nil {
		return nil, errors.New("error to create oauth service")
	}

	tokenInfo, err := oauth2Service.Tokeninfo().IdToken(token).Do()

	if err != nil {
		return nil, errors.New("invalid token info")
	}

	return tokenInfo, nil
}
