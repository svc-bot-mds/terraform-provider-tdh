package upgrade_service

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

// UpdateClusterVersion updates the version of the TDH cluster
func (s *Service) UpdateClusterVersion(id string, requestBody *UpdateClusterVersionRequest) (*model.UpdateClusterVersionResponse, error) {
	urlPath := fmt.Sprintf("%s/%s", s.Endpoint, Upgrade)
	var response model.UpdateClusterVersionResponse

	_, err := s.Api.Post(&urlPath, requestBody, &response)
	if err != nil {
		return &response, err
	}

	return &response, nil
}
