package upgrade_service

import (
	"fmt"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/model"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/core"
	"strings"
)

const (
	EndPoint = "upgradeservice"
)

type Service struct {
	*core.Service
}

func NewService(hostUrl *string, root *core.Root) *Service {
	return &Service{
		Service: core.NewService(hostUrl, EndPoint, root),
	}
}

// GetClusterTargetVersions - Returns all the versions available for upgrading a cluster
func (s *Service) GetClusterTargetVersions(clusterId string) (*model.ClusterTargetVersionsResponse, error) {
	if strings.TrimSpace(clusterId) == "" {
		return nil, fmt.Errorf("clusterId cannot be empty")
	}
	urlPath := fmt.Sprintf("%s/%s/%s/%s", s.Endpoint, Upgrade, clusterId, TargetVersions)
	var response model.ClusterTargetVersionsResponse

	_, err := s.Api.Get(&urlPath, nil, &response)
	if err != nil {
		return &response, err
	}

	return &response, nil
}

// UpdateClusterVersion updates the version of the TDH cluster
func (s *Service) UpdateClusterVersion(requestBody *UpdateClusterVersionRequest) (*model.UpdateClusterVersionResponse, error) {
	urlPath := fmt.Sprintf("%s/%s", s.Endpoint, Upgrade)
	var response model.UpdateClusterVersionResponse

	_, err := s.Api.Post(&urlPath, requestBody, &response)
	if err != nil {
		return &response, err
	}

	return &response, nil
}
