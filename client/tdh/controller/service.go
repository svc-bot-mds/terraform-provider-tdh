package controller

import (
	"fmt"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/model"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/core"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/utils"
	"strings"
)

const (
	EndPoint = "controller"
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

// GetClusters - Returns page of clusters
func (s *Service) GetClusters(query *ClustersQuery) (model.Paged[model.Cluster], error) {
	urlPath := fmt.Sprintf("%s/%s", s.Endpoint, Clusters)
	var response model.Paged[model.Cluster]

	if query.Size == 0 {
		query.Size = defaultPage.Size
	}

	_, err := s.Api.Get(&urlPath, query, &response)
	if err != nil {
		return response, err
	}

	return response, nil
}

// GetAllClusters - Returns list of all clusters
func (s *Service) GetAllClusters(query *ClustersQuery) ([]model.Cluster, error) {
	var clusters []model.Cluster
	for {
		queriedClusters, err := s.GetClusters(query)
		if err != nil {
			return clusters, err
		}
		clusters = append(clusters, *queriedClusters.Get()...)
		nextPage := utils.GetNextPageInfo(queriedClusters.GetPage())
		if nextPage == nil {
			break
		}
		query.PageQuery = *nextPage
	}
	return clusters, nil
}

// GetCluster - Returns the cluster by ID
func (s *Service) GetCluster(id string) (*model.Cluster, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("ID cannot be empty")
	}
	urlPath := fmt.Sprintf("%s/%s/%s", s.Endpoint, Clusters, id)
	var response model.Cluster

	_, err := s.Api.Get(&urlPath, nil, &response)
	if err != nil {
		return &response, err
	}

	return &response, err
}

// CreateCluster - Submits a request to create cluster
func (s *Service) CreateCluster(requestBody *ClusterCreateRequest) (*model.TaskResponse, error) {
	if requestBody == nil {
		return nil, fmt.Errorf("requestBody cannot be nil")
	}
	urlPath := fmt.Sprintf("%s/%s", s.Endpoint, Clusters)
	var response model.TaskResponse

	_, err := s.Api.Post(&urlPath, requestBody, &response)
	if err != nil {
		return &response, err
	}

	return &response, nil
}

// UpdateCluster - Submits a request to update cluster
func (s *Service) UpdateCluster(id string, requestBody *ClusterUpdateRequest) (*model.Cluster, error) {
	if id == "" {
		return nil, fmt.Errorf("cluster ID cannot be empty")
	}
	if requestBody == nil {
		return nil, fmt.Errorf("requestBody cannot be nil")
	}
	urlPath := fmt.Sprintf("%s/%s/%s", s.Endpoint, Clusters, id)
	var response model.Cluster

	_, err := s.Api.Patch(&urlPath, requestBody.Tags, &response)
	if err != nil {
		return &response, err
	}

	return &response, nil
}

// UpdateClusterNetworkPolicies - Submits a request to update cluster network policies
func (s *Service) UpdateClusterNetworkPolicies(id string, requestBody *ClusterNetworkPoliciesUpdateRequest) ([]byte, error) {
	if id == "" {
		return nil, fmt.Errorf("cluster ID cannot be empty")
	}
	if requestBody == nil {
		return nil, fmt.Errorf("requestBody cannot be nil")
	}
	urlPath := fmt.Sprintf("%s/%s/%s/%s", s.Endpoint, Clusters, id, NetworkPolicy)

	bodyBytes, err := s.Api.Patch(&urlPath, requestBody, nil)
	if err != nil {
		return nil, err
	}

	return bodyBytes, nil
}

// DeleteCluster - Submits a request to delete cluster
func (s *Service) DeleteCluster(id string) (*model.TaskResponse, error) {
	urlPath := fmt.Sprintf("%s/%s/%s", s.Endpoint, Clusters, id)
	var response model.TaskResponse

	_, err := s.Api.Delete(&urlPath, nil, &response)
	if err != nil {
		return &response, err
	}

	return &response, nil
}

// GetServiceInstanceTypes - Returns list of clusters
func (s *Service) GetServiceInstanceTypes(serviceTypeQuery *InstanceTypesQuery) (model.InstanceTypeList, error) {
	reqUrl := fmt.Sprintf("%s/%s/%s", s.Endpoint, Services, InstanceTypes)
	var response model.InstanceTypeList

	if serviceTypeQuery.Size == 0 {
		serviceTypeQuery.Size = defaultPage.Size
	}

	_, err := s.Api.Get(&reqUrl, serviceTypeQuery, &response)
	if err != nil {
		return response, err
	}

	return response, nil
}

// GetClusterMetaData - Returns the cluster metadata by ID
func (s *Service) GetClusterMetaData(id string) (*model.ClusterMetaData, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("ID cannot be empty")
	}
	urlPath := fmt.Sprintf("%s/%s/%s/%s", s.Endpoint, Clusters, id, MetaData)
	var response model.ClusterMetaData

	_, err := s.Api.Get(&urlPath, nil, &response)
	if err != nil {
		return &response, err
	}

	return &response, err
}

func (s *Service) GetClusterCountByService() ([]model.ClusterCountByService, error) {
	var response []model.ClusterCountByService

	reqUrl := fmt.Sprintf("%s/%s/%s/%s", s.Endpoint, FleetManagement, SRE_cluster, Count)

	_, err := s.Api.Get(&reqUrl, nil, &response)
	if err != nil {
		return response, err
	}
	return response, nil
}

func (s *Service) GetResourceByService() ([]model.ResourceByService, error) {
	var response []model.ResourceByService

	reqUrl := fmt.Sprintf("%s/%s/%s/%s", s.Endpoint, FleetManagement, SRE_cluster, ResourceByService)

	_, err := s.Api.Get(&reqUrl, nil, &response)
	if err != nil {
		return response, err
	}
	return response, nil
}

func (s *Service) GetFleetDetails(query *FleetsQuery) (model.Paged[model.SreCustomerInfo], error) {
	var response model.Paged[model.SreCustomerInfo]

	reqUrl := fmt.Sprintf("%s/%s", s.Endpoint, Mdsfleets)

	_, err := s.Api.Get(&reqUrl, query, &response)
	if err != nil {
		return response, err
	}
	return response, nil

}
