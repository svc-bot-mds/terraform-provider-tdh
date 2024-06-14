package auth

import (
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/constants/oauth_type"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/model"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/core"
)

const (
	EndPoint = "authservice"
)

type Service struct {
	*core.Service
}

func NewService(hostUrl *string, root *core.Root) *Service {
	return &Service{
		Service: core.NewService(hostUrl, EndPoint, root),
	}
}

// GetAccessToken - Get a new token for user
func (s *Service) GetAccessToken() (*TokenResponse, error) {
	if s.Api.AuthToUse.ApiToken == "" && s.Api.AuthToUse.OAuthAppType == oauth_type.ApiToken {
		return nil, fmt.Errorf("define API Token")
	}

	if s.Api.AuthToUse.ClientId == "" && s.Api.AuthToUse.OAuthAppType == oauth_type.ClientCredentials {
		return nil, fmt.Errorf("define TDH Client Id")
	}
	if s.Api.AuthToUse.ClientSecret == "" && s.Api.AuthToUse.OAuthAppType == oauth_type.ClientCredentials {
		return nil, fmt.Errorf("define TDH Client Secret")
	}
	if s.Api.AuthToUse.OrgId == "" && s.Api.AuthToUse.OAuthAppType == oauth_type.ClientCredentials {
		return nil, fmt.Errorf("define TDH Org Id")
	}
	if s.Api.AuthToUse.Username == "" && s.Api.AuthToUse.OAuthAppType == oauth_type.UserCredentials {
		return nil, fmt.Errorf("define TDH Username Credentials")
	}
	if s.Api.AuthToUse.Password == "" && s.Api.AuthToUse.OAuthAppType == oauth_type.UserCredentials {
		return nil, fmt.Errorf("define TDH Password")
	}

	reqUrl := fmt.Sprintf("%s/%s", s.Endpoint, Token)

	tokenRequest := TokenRequest{
		ApiKey:        s.Api.AuthToUse.ApiToken,
		ClientId:      s.Api.AuthToUse.ClientId,
		ClientSecret:  s.Api.AuthToUse.ClientSecret,
		AccessToken:   s.Api.AuthToUse.AccessToken,
		OAuthAppTypes: s.Api.AuthToUse.OAuthAppType,
		OrgId:         s.Api.AuthToUse.OrgId,
		Username:      s.Api.AuthToUse.Username,
		Password:      s.Api.AuthToUse.Password,
	}
	if s.Api.AuthToUse.OAuthAppType == oauth_type.ClientCredentials {
		s.Api.OrgId = s.Api.AuthToUse.OrgId
	}
	body, err := s.Api.Post(&reqUrl, &tokenRequest, nil)
	if err != nil {
		return nil, err
	}

	ar := TokenResponse{
		Token: string(body),
	}

	err = s.processAuthResponse(&ar)
	if err != nil {
		return nil, err
	}

	return &ar, nil
}

func (s *Service) processAuthResponse(response *TokenResponse) error {
	s.Api.Token = &response.Token
	token, err := jwt.Parse(*s.Api.Token, nil)
	if token == nil {
		return err
	}
	if s.Api.AuthToUse.OAuthAppType == oauth_type.ApiToken {
		claims, _ := token.Claims.(jwt.MapClaims)

		s.Api.OrgId = claims["context_name"].(string)
	}

	return nil
}

func (s *Service) GetSmtpDetails() (model.Smtp, error) {
	var response model.Smtp

	reqUrl := fmt.Sprintf("%s/%s", s.Endpoint, SMTP)

	_, err := s.Api.Get(&reqUrl, nil, &response)
	if err != nil {
		return response, err
	}
	return response, nil
}


func (s *Service) CreateSmtpDetails(requestBody SmtpRequest) (model.Smtp, error) {
	var response model.Smtp

	reqUrl := fmt.Sprintf("%s/%s", s.Endpoint, SMTP)

	_, err := s.Api.Post(&reqUrl, requestBody, &response)
	if err != nil {
		return response, err
	}
	return response, nil
}


func (s *Service) UpdateSmtpDetails(requestBody SmtpRequest) (model.Smtp, error) {
	var response model.Smtp

	reqUrl := fmt.Sprintf("%s/%s", s.Endpoint, SMTP)

	_, err := s.Api.Patch(&reqUrl, requestBody, &response)
	if err != nil {
		return response, err
	}
	return response, nil
}
