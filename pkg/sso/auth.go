package sso

import (
	"context"
	"errors"
	"google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
	"net/http"
)

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
