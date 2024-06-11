package service_metadata

import (
	"fmt"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/model"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/core"
)

var (
	defaultPage = &model.PageQuery{
		Index: 0,
		Size:  100,
	}
)

const (
	EndPoint = "servicemetadata"
)

type Service struct {
	*core.Service
}

func NewService(hostUrl *string, root *core.Root) *Service {
	return &Service{
		Service: core.NewService(hostUrl, EndPoint, root),
	}
}

func (s *Service) GetNetworkPorts() ([]model.NetworkPorts, error) {
	reqUrl := fmt.Sprintf("%s/%s/%s", s.Endpoint, MdsServices, NetworkPorts)

	var response []model.NetworkPorts

	_, err := s.Api.Get(&reqUrl, nil, &response)
	if err != nil {
		return response, err
	}

	return response, nil
}

// GetRoles - Return list of Roles for the users
func (s *Service) GetRoles(query *RolesQuery) (model.Roles, error) {
	reqUrl := fmt.Sprintf("%s/%s/%s", s.Endpoint, MdsServices, Roles)
	var response model.Roles

	if query.Size == 0 {
		query.Size = defaultPage.Size
	}

	_, err := s.Api.Get(&reqUrl, query, &response)
	if err != nil {
		return response, err
	}
	return response, nil
}

// GetServiceRoles - Return list of Roles for the service
func (s *Service) GetServiceRoles(query *RolesQuery) (model.ServiceRoles, error) {
	reqUrl := fmt.Sprintf("%s/%s/%s", s.Endpoint, MdsServices, Roles)
	var response model.ServiceRoles

	if query.Size == 0 {
		query.Size = defaultPage.Size
	}

	_, err := s.Api.Get(&reqUrl, query, &response)
	if err != nil {
		return response, err
	}
	return response, nil
}

// GetPolicyTypes - Returns the policy types
func (s *Service) GetPolicyTypes() ([]string, error) {
	urlPath := fmt.Sprintf("%s/%s/%s/%s", s.Endpoint, MdsServices, Policies, Types)
	var response []string

	_, err := s.Api.Get(&urlPath, nil, &response)
	if err != nil {
		return response, err
	}

	return response, err
}
