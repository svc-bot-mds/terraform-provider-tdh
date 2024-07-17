package customer_metadata

import (
	"fmt"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/constants/account_type"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/model"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/core"
	"strings"
)

const (
	EndPoint = "customermetadata"
)

var (
	defaultPage = &model.PageQuery{
		Index: 0,
		Size:  100,
	}
)

type Service struct {
	*core.Service
}

func NewService(hostUrl *string, root *core.Root) *Service {
	return &Service{
		Service: core.NewService(hostUrl, EndPoint, root),
	}
}

// GetPolicies - Returns list of Policies
func (s *Service) GetPolicies(query *PoliciesQuery) (model.Paged[model.Policy], error) {
	reqUrl := fmt.Sprintf("%s/%s", s.Endpoint, Policies)
	var response model.Paged[model.Policy]

	if query.Size == 0 {
		query.Size = defaultPage.Size
	}

	_, err := s.Api.Get(&reqUrl, query, &response)
	if err != nil {
		return response, err
	}
	return response, nil
}

// GetUsers - Return list of Users
func (s *Service) GetUsers(query *UsersQuery) (model.Paged[model.User], error) {
	var response model.Paged[model.User]
	if query == nil {
		return response, fmt.Errorf("query cannot be nil")
	}
	query.AccountType = account_type.USER_ACCOUNT
	var reqUrl string
	if s.Api.IsSre {
		reqUrl = fmt.Sprintf("%s/%s", s.Endpoint, TdhUsers)
	} else {
		reqUrl = fmt.Sprintf("%s/%s", s.Endpoint, Users)
	}

	if query.Size == 0 {
		query.Size = defaultPage.Size
	}

	_, err := s.Api.Get(&reqUrl, query, &response)
	if err != nil {
		return response, err
	}
	return response, nil
}

// CreateUser - Submits a request to create user
func (s *Service) CreateUser(requestBody *CreateUserRequest) error {
	if requestBody == nil {
		return fmt.Errorf("requestBody cannot be nil")
	}
	requestBody.AccountType = account_type.USER_ACCOUNT
	var reqUrl string
	if s.Api.IsSre {
		reqUrl = fmt.Sprintf("%s/%s", s.Endpoint, TdhUsers)
	} else {
		reqUrl = fmt.Sprintf("%s/%s", s.Endpoint, Users)
	}

	_, err := s.Api.Post(&reqUrl, requestBody, nil)
	if err != nil {
		return err
	}

	return nil
}

// UpdateUser - Submits a request to update user
func (s *Service) UpdateUser(id string, requestBody *UserUpdateRequest) error {
	if id == "" {
		return fmt.Errorf("user ID cannot be empty")
	}
	if requestBody == nil {
		return fmt.Errorf("requestBody cannot be nil")
	}
	urlPath := fmt.Sprintf("%s/%s/%s", s.Endpoint, Users, id)

	_, err := s.Api.Patch(&urlPath, requestBody, nil)
	return err
}

// GetUser - Returns the user by ID
func (s *Service) GetUser(id string) (*model.User, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("ID cannot be empty")
	}
	urlPath := fmt.Sprintf("%s/%s/%s", s.Endpoint, Users, id)
	var response model.User

	_, err := s.Api.Get(&urlPath, nil, &response)
	if err != nil {
		return &response, err
	}

	return &response, err
}

// DeleteUser - Submits a request to delete user
func (s *Service) DeleteUser(id string, query *DeleteUserQuery) error {
	urlPath := fmt.Sprintf("%s/%s/%s", s.Endpoint, Users, id)

	if query == nil {
		query.DeleteFromIdp = false
	}
	_, err := s.Api.Delete(&urlPath, query, nil)
	if err != nil {
		return err
	}

	return nil
}

// GetServiceAccounts - Return list of Service Accounts
func (s *Service) GetServiceAccounts(query *ServiceAccountsQuery) (model.Paged[model.ServiceAccount], error) {

	var response model.Paged[model.ServiceAccount]
	if query == nil {
		return response, fmt.Errorf("query cannot be nil")
	}

	query.AccountType = account_type.SERVICE_ACCOUNT
	reqUrl := fmt.Sprintf("%s/%s", s.Endpoint, Users)

	if query.Size == 0 {
		query.Size = defaultPage.Size
	}

	_, err := s.Api.Get(&reqUrl, query, &response)
	if err != nil {
		return response, err
	}

	return response, nil
}

// CreateServiceAccount - Submits a request to create service account
func (s *Service) CreateServiceAccount(requestBody *CreateSvcAccountRequest) (*model.ServiceAccountCreate, error) {
	if requestBody == nil {
		return nil, fmt.Errorf("requestBody cannot be nil")
	}
	var response model.ServiceAccountCreate
	requestBody.AccountType = account_type.SERVICE_ACCOUNT

	urlPath := fmt.Sprintf("%s/%s", s.Endpoint, Users)

	_, err := s.Api.Post(&urlPath, requestBody, &response)
	if err != nil {
		return &response, err
	}

	return &response, err
}

// GetServiceAccountOauthApp - Fetch oauthDetails for the service account
func (s *Service) GetServiceAccountOauthApp(id string) (*model.ServiceAccountOauthApp, error) {

	var response model.ServiceAccountOauthApp

	urlPath := fmt.Sprintf("%s/%s/%s/%s", s.Endpoint, Users, id, OAuthApps)
	_, err := s.Api.Get(&urlPath, nil, &response)
	if err != nil {
		return &response, err
	}

	return &response, err
}

// UpdateServiceAccountOauthApp - To Update the Oauth app details
func (s *Service) UpdateServiceAccountOauthApp(id string, requestBody *OauthAppUpdateRequest, appId string) (*model.ServiceAccountOauthApp, error) {

	var response model.ServiceAccountOauthApp

	urlPath := fmt.Sprintf("%s/%s/%s/%s/%s", s.Endpoint, Users, id, OAuthApps, appId)
	_, err := s.Api.Patch(&urlPath, requestBody, &response)

	if err != nil {
		return &response, err
	}

	return &response, err
}

// UpdateServiceAccount - Submits a request to update service account
func (s *Service) UpdateServiceAccount(id string, requestBody *SvcAccountUpdateRequest) error {
	if id == "" {
		return fmt.Errorf("service account ID cannot be empty")
	}
	if requestBody == nil {
		return fmt.Errorf("requestBody cannot be nil")
	}
	urlPath := fmt.Sprintf("%s/%s/%s", s.Endpoint, Users, id)
	_, err := s.Api.Patch(&urlPath, requestBody, nil)
	return err
}

// GetServiceAccount - Returns the service account by ID
func (s *Service) GetServiceAccount(id string) (*model.ServiceAccount, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("ID cannot be empty")
	}
	urlPath := fmt.Sprintf("%s/%s/%s", s.Endpoint, Users, id)
	var response model.ServiceAccount

	_, err := s.Api.Get(&urlPath, nil, &response)
	if err != nil {
		return &response, err
	}

	return &response, err
}

// DeleteServiceAccount - Submits a request to delete service account
func (s *Service) DeleteServiceAccount(id string) error {
	urlPath := fmt.Sprintf("%s/%s/%s", s.Endpoint, Users, id)

	_, err := s.Api.Delete(&urlPath, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

// CreatePolicy - Submits a request to create policy
func (s *Service) CreatePolicy(requestBody *CreateUpdatePolicyRequest) (*model.Policy, error) {
	if requestBody == nil {
		return nil, fmt.Errorf("requestBody cannot be nil")
	}
	var response model.Policy
	urlPath := fmt.Sprintf("%s/%s", s.Endpoint, Policies)

	_, err := s.Api.Post(&urlPath, requestBody, &response)
	if err != nil {
		return &response, err
	}

	return &response, err
}

// UpdatePolicy - Submits a request to update policy
func (s *Service) UpdatePolicy(id string, requestBody *CreateUpdatePolicyRequest) (*model.Policy, error) {
	if id == "" {
		return nil, fmt.Errorf("policy ID cannot be empty")
	}
	if requestBody == nil {
		return nil, fmt.Errorf("requestBody cannot be nil")
	}
	urlPath := fmt.Sprintf("%s/%s/%s", s.Endpoint, Policies, id)
	var response model.Policy
	_, err := s.Api.Put(&urlPath, requestBody, &response)
	return &response, err
}

// GetPolicy - Submits a request to fetch policy
func (s *Service) GetPolicy(id string) (*model.Policy, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("ID cannot be empty")
	}
	urlPath := fmt.Sprintf("%s/%s/%s", s.Endpoint, Policies, id)
	var response model.Policy

	_, err := s.Api.Get(&urlPath, nil, &response)
	if err != nil {
		return &response, err
	}

	return &response, err
}

// DeletePolicy - Submits a request to delete policy
func (s *Service) DeletePolicy(id string) error {
	urlPath := fmt.Sprintf("%s/%s/%s", s.Endpoint, Policies, id)

	_, err := s.Api.Delete(&urlPath, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

// GetLocalUsers - Return list of Local Users
func (s *Service) GetLocalUsers(query *LocalUsersQuery) (model.Paged[model.LocalUser], error) {
	var response model.Paged[model.LocalUser]
	if query == nil {
		return response, fmt.Errorf("query cannot be nil")
	}

	reqUrl := fmt.Sprintf("%s/%s", s.Endpoint, LocalUsers)

	if query.Size == 0 {
		query.Size = defaultPage.Size
	}

	_, err := s.Api.Get(&reqUrl, query, &response)
	if err != nil {
		return response, err
	}
	return response, nil
}

// CreateLocalUser - Submits a request to create local ser
func (s *Service) CreateLocalUser(requestBody *CreateLocalUserRequest) (*[]model.TaskResponse, error) {
	if requestBody == nil {
		return nil, fmt.Errorf("requestBody cannot be nil")
	}
	urlPath := fmt.Sprintf("%s/%s", s.Endpoint, LocalUsers)

	var response []model.TaskResponse
	_, err := s.Api.Post(&urlPath, requestBody, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// UpdateLocalUser - Submits a request to update local user
func (s *Service) UpdateLocalUser(id string, requestBody *LocalUserUpdateRequest) (*[]model.TaskResponse, error) {
	if id == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}
	if requestBody == nil {
		return nil, fmt.Errorf("requestBody cannot be nil")
	}
	urlPath := fmt.Sprintf("%s/%s/%s", s.Endpoint, LocalUsers, id)

	var response []model.TaskResponse
	_, err := s.Api.Patch(&urlPath, requestBody, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// GetLocalUser - Returns the local user by ID
func (s *Service) GetLocalUser(id string) (*model.LocalUser, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("ID cannot be empty")
	}
	urlPath := fmt.Sprintf("%s/%s/%s", s.Endpoint, LocalUsers, id)
	var response model.LocalUser

	_, err := s.Api.Get(&urlPath, nil, &response)
	if err != nil {
		return &response, err
	}

	return &response, err
}

// DeleteLocalUser - Submits a request to delete local user
func (s *Service) DeleteLocalUser(id string) (*[]model.TaskResponse, error) {
	urlPath := fmt.Sprintf("%s/%s/%s", s.Endpoint, LocalUsers, id)

	var response []model.TaskResponse
	_, err := s.Api.Delete(&urlPath, nil, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
