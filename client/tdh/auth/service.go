package auth

import (
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/constants/oauth_type"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/model"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/core"
)

const (
	endPoint = "authservice"
)

type Service struct {
	*core.Service
}

func NewService(hostUrl *string, root *core.Root) *Service {
	return &Service{
		Service: core.NewService(hostUrl, endPoint, root),
	}
}

// GetAccessToken - Get a new token for user
func (s *Service) GetAccessToken() (*TokenResponse, error) {
	fmt.Println("Going to grab auth token")
	if s.Api.AuthToUse.OAuthAppType == oauth_type.ClientCredentials {
		s.Api.OrgId = s.Api.AuthToUse.OrgId
	}

	authToUse := s.Api.AuthToUse
	if authToUse.ApiToken == "" && authToUse.OAuthAppType == oauth_type.ApiToken {
		return nil, fmt.Errorf("define API Token")
	}

	if authToUse.ClientId == "" && authToUse.OAuthAppType == oauth_type.ClientCredentials {
		return nil, fmt.Errorf("define TDH Client Id")
	}
	if authToUse.ClientSecret == "" && authToUse.OAuthAppType == oauth_type.ClientCredentials {
		return nil, fmt.Errorf("define TDH Client Secret")
	}
	if authToUse.OrgId == "" && authToUse.OAuthAppType == oauth_type.ClientCredentials {
		return nil, fmt.Errorf("define TDH Org Id")
	}
	if authToUse.Username == "" && authToUse.OAuthAppType == oauth_type.UserCredentials {
		return nil, fmt.Errorf("define TDH Username")
	}
	if authToUse.Password == "" && authToUse.OAuthAppType == oauth_type.UserCredentials {
		return nil, fmt.Errorf("define TDH Password")
	}
	if authToUse.OrgId == "" && authToUse.OAuthAppType == oauth_type.UserCredentials {
		return nil, fmt.Errorf("define TDH Org Id")
	}

	reqUrl := fmt.Sprintf("%s/%s", s.Endpoint, Token)
	tokenRequest := TokenRequest{
		ApiKey:        authToUse.ApiToken,
		ClientId:      authToUse.ClientId,
		ClientSecret:  authToUse.ClientSecret,
		AccessToken:   authToUse.AccessToken,
		OAuthAppTypes: authToUse.OAuthAppType,
		OrgId:         authToUse.OrgId,
		Username:      authToUse.Username,
		Password:      authToUse.Password,
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
	claims, _ := token.Claims.(jwt.MapClaims)
	var perms = claims["perms"]
	// Check if the target string StgManagedDataService:SRE exists in the response array
	found := false
	for _, item := range perms.([]interface{}) {
		if item == "StgManagedDataService:SRE" {
			found = true
			break
		}
	}

	if found {
		s.Api.IsSre = true
	} else {
		s.Api.IsSre = false
	}

	if s.Api.AuthToUse.OAuthAppType == oauth_type.ApiToken {
		s.Api.OrgId = claims["context_name"].(string)
	}

	return nil
}

// Login - Logs in user and return cookies
func (s *Service) Login() error {
	fmt.Println("Trying login")
	if s.Api.AuthToUse.OAuthAppType != oauth_type.UserCredentials {
		return nil
	}

	authToUse := s.Api.AuthToUse
	if authToUse.Username == "" && authToUse.OAuthAppType == oauth_type.UserCredentials {
		return fmt.Errorf("define TDH Username")
	}
	if authToUse.Password == "" && authToUse.OAuthAppType == oauth_type.UserCredentials {
		return fmt.Errorf("define TDH Password")
	}

	reqUrl := fmt.Sprintf("%s/%s/%s", s.Endpoint, Auth, Login)
	tokenRequest := TokenRequest{
		Username: authToUse.Username,
		Password: authToUse.Password,
	}
	_, err := s.Api.Post(&reqUrl, &tokenRequest, nil)
	if err != nil {
		return err
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
