package infra_connector

import (
	"fmt"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/model"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/core"
	"strings"
)

const (
	EndPoint = "infra-connector"
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

func (s *Service) GetRegionsWithDataPlanes(regionsQuery *DataPlaneRegionsQuery) (map[string][]string, error) {
	reqUrl := fmt.Sprintf("%s/%s/%s", s.Endpoint, K8sCluster, Resource)

	var response map[string][]string

	_, err := s.Api.Get(&reqUrl, regionsQuery, &response)
	if err != nil {
		return response, err
	}

	return response, nil
}

func (s *Service) GetCloudAccounts(query *CloudAccountsQuery) (model.Paged[model.CloudAccount], error) {
	var response model.Paged[model.CloudAccount]
	if query == nil {
		return response, fmt.Errorf("query cannot be nil")
	}

	reqUrl := fmt.Sprintf("%s/%s/%s", s.Endpoint, Internal, CloudAccount)

	if query.Size == 0 {
		query.Size = defaultPage.Size
	}

	_, err := s.Api.Get(&reqUrl, query, &response)
	if err != nil {
		return response, err
	}
	return response, nil
}

// GetCloudAccount - Submits a request to fetch cloud account
func (s *Service) GetCloudAccount(id string) (*model.CloudAccount, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("ID cannot be empty")
	}
	urlPath := fmt.Sprintf("%s/%s/%s", s.Endpoint, CloudAccount, id)
	var response model.CloudAccount

	_, err := s.Api.Get(&urlPath, nil, &response)
	if err != nil {
		return &response, err
	}

	return &response, err
}

func (s *Service) GetCertificates(query *CertificateQuery) (model.Paged[model.Certificate], error) {
	var response model.Paged[model.Certificate]
	if query == nil {
		return response, fmt.Errorf("query cannot be nil")
	}

	reqUrl := fmt.Sprintf("%s/%s/%s", s.Endpoint, Internal, Certificate)

	if query.Size == 0 {
		query.Size = defaultPage.Size
	}

	_, err := s.Api.Get(&reqUrl, query, &response)
	if err != nil {
		return response, err
	}
	return response, nil
}

func (s *Service) GetDnsconfig(query *DNSQuery) (model.Paged[model.Dns], error) {
	var response model.Paged[model.Dns]
	if query == nil {
		return response, fmt.Errorf("query cannot be nil")
	}

	reqUrl := fmt.Sprintf("%s/%s", s.Endpoint, DNSConfig)

	if query.Size == 0 {
		query.Size = defaultPage.Size
	}

	_, err := s.Api.Get(&reqUrl, query, &response)
	if err != nil {
		return response, err
	}
	return response, nil
}

func (s *Service) GetTshirtSizes(query *TshirtSizesQuery) (model.Paged[model.TshirtSize], error) {
	var response model.Paged[model.TshirtSize]
	if query == nil {
		return response, fmt.Errorf("query cannot be nil")
	}

	reqUrl := fmt.Sprintf("%s/%s/%s", s.Endpoint, K8sCluster, TshirtSize)

	if query.Size == 0 {
		query.Size = defaultPage.Size
	}

	_, err := s.Api.Get(&reqUrl, query, &response)
	if err != nil {
		return response, err
	}
	return response, nil
}

func (s *Service) GetProviderTypes() ([]string, error) {
	urlPath := fmt.Sprintf("%s/%s/%s", s.Endpoint, CloudAccount, Types)
	var response []string

	_, err := s.Api.Get(&urlPath, nil, &response)
	if err != nil {
		return response, err
	}

	return response, err
}

func (s *Service) GetDataPlaneRegions() ([]model.DataPlaneRegion, error) {
	var response []model.DataPlaneRegion

	reqUrl := fmt.Sprintf("%s/%s", s.Endpoint, CloudProviders)

	_, err := s.Api.Get(&reqUrl, nil, &response)
	if err != nil {
		return response, err
	}
	return response, nil
}

// CreateDataPlane - Submits a request to create dataplane
func (s *Service) CreateDataPlane(requestBody *DataPlaneCreateRequest) (*model.TaskResponse, error) {
	if requestBody == nil {
		return nil, fmt.Errorf("requestBody cannot be nil")
	}
	var response model.TaskResponse
	urlPath := fmt.Sprintf("%s/%s/%s/%s", s.Endpoint, Internal, K8sCluster, DataplaneOnboard)

	_, err := s.Api.Post(&urlPath, requestBody, &response)
	if err != nil {
		return &response, err
	}

	return &response, err
}

func (s *Service) UpdateDataPlane(id string, requestBody *DataPlaneUpdateRequest) error {
	if requestBody == nil {
		return fmt.Errorf("requestBody cannot be nil")
	}
	urlPath := fmt.Sprintf("%s/%s/%s/%s", s.Endpoint, Internal, K8sCluster, id)

	_, err := s.Api.Patch(&urlPath, requestBody, nil)
	if err != nil {
		return err
	}

	return err
}

func (s *Service) GetDataPlanes(query *DataPlaneQuery) (model.Paged[model.DataPlane], error) {
	urlPath := fmt.Sprintf("%s/%s/%s", s.Endpoint, Internal, K8sCluster)
	var response model.Paged[model.DataPlane]

	if query.Size == 0 {
		query.Size = defaultPage.Size
	}

	_, err := s.Api.Get(&urlPath, query, &response)
	if err != nil {
		return response, err
	}

	return response, nil
}

func (s *Service) GetDataPlaneById(id string) (model.DataPlane, error) {
	urlPath := fmt.Sprintf("%s/%s/%s/%s", s.Endpoint, Internal, K8sCluster, id)
	var response model.DataPlane

	_, err := s.Api.Get(&urlPath, nil, &response)
	if err != nil {
		return response, err
	}

	return response, nil
}

// DeleteDataPlane - Submits a request to delete dataplane
func (s *Service) DeleteDataPlane(id string) error {
	urlPath := fmt.Sprintf("%s/%s/%s/%s/%s", s.Endpoint, Internal, K8sCluster, DataplaneOnboard, id)

	_, err := s.Api.Delete(&urlPath, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) CreateCloudAccount(requestBody *CloudAccountCreateRequest) (*model.CloudAccount, error) {
	if requestBody == nil {
		return nil, fmt.Errorf("requestBody cannot be nil")
	}
	var response model.CloudAccount
	urlPath := fmt.Sprintf("%s/%s/%s", s.Endpoint, Internal, CloudAccount)

	_, err := s.Api.Post(&urlPath, requestBody, &response)
	if err != nil {
		return &response, err
	}

	return &response, err
}

// UpdateCloudAccount - To Update the cloud account
func (s *Service) UpdateCloudAccount(id string, requestBody *CloudAccountUpdateRequest) error {

	urlPath := fmt.Sprintf("%s/%s/%s", s.Endpoint, CloudAccount, id)
	_, err := s.Api.Put(&urlPath, requestBody, nil)

	if err != nil {
		return err
	}

	return err
}

// DeleteCloudAccount - Submits a request to delete cloud account
func (s *Service) DeleteCloudAccount(id string) error {
	urlPath := fmt.Sprintf("%s/%s/%s", s.Endpoint, CloudAccount, id)

	_, err := s.Api.Delete(&urlPath, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) CreateCertificate(requestBody *CertificateCreateRequest) (*model.Certificate, error) {
	if requestBody == nil {
		return nil, fmt.Errorf("requestBody cannot be nil")
	}
	var response model.Certificate
	urlPath := fmt.Sprintf("%s/%s/%s", s.Endpoint, Internal, Certificate)

	_, err := s.Api.Post(&urlPath, requestBody, &response)
	if err != nil {
		return &response, err
	}

	return &response, err
}

func (s *Service) UpdateCertificate(id string, requestBody *CertificateUpdateRequest) (*model.Certificate, error) {
	if requestBody == nil {
		return nil, fmt.Errorf("requestBody cannot be nil")
	}
	var response model.Certificate
	urlPath := fmt.Sprintf("%s	/%s/%s", s.Endpoint, Certificate, id)

	_, err := s.Api.Post(&urlPath, requestBody, &response)
	if err != nil {
		return &response, err
	}

	return &response, err
}

func (s *Service) GetCertificate(id string) (model.Certificate, error) {
	var response model.Certificate

	reqUrl := fmt.Sprintf("%s/%s/%s", s.Endpoint, Certificate, id)

	_, err := s.Api.Get(&reqUrl, nil, &response)
	if err != nil {
		return response, err
	}
	return response, err
}

// DeleteCertificate - Submits a request to delete certificate
func (s *Service) DeleteCertificate(id string) error {
	urlPath := fmt.Sprintf("%s/%s/%s", s.Endpoint, Certificate, id)

	_, err := s.Api.Delete(&urlPath, nil, nil)
	if err != nil {
		return err
	}

	return err
}

func (s *Service) GetObjectStorages(query *ObjectStorageQuery) (model.Paged[model.ObjectStorage], error) {
	var response model.Paged[model.ObjectStorage]
	if query == nil {
		return response, fmt.Errorf("query cannot be nil")
	}

	reqUrl := fmt.Sprintf("%s/%s", s.Endpoint, ObjectStore)

	if query.Size == 0 {
		query.Size = defaultPage.Size
	}

	_, err := s.Api.Get(&reqUrl, query, &response)
	if err != nil {
		return response, err
	}
	return response, nil
}

func (s *Service) GetObjectStorage(id string) (model.ObjectStorage, error) {
	var response model.ObjectStorage

	reqUrl := fmt.Sprintf("%s/%s/%s", s.Endpoint, ObjectStore, id)

	_, err := s.Api.Get(&reqUrl, nil, &response)
	if err != nil {
		return response, err
	}
	return response, err
}

func (s *Service) CreateObjectStorage(requestBody *ObjectStorageCreateRequest) (*model.ObjectStorage, error) {
	if requestBody == nil {
		return nil, fmt.Errorf("requestBody cannot be nil")
	}
	var response model.ObjectStorage
	urlPath := fmt.Sprintf("%s/%s", s.Endpoint, ObjectStore)

	_, err := s.Api.Post(&urlPath, requestBody, &response)
	if err != nil {
		return &response, err
	}

	return &response, err
}

func (s *Service) UpdateObjectStore(id string, requestBody *ObjectStorageUpdateRequest) (*model.ObjectStorage, error) {
	if requestBody == nil {
		return nil, fmt.Errorf("requestBody cannot be nil")
	}
	var response model.ObjectStorage
	urlPath := fmt.Sprintf("%s	/%s/%s", s.Endpoint, ObjectStore, id)

	_, err := s.Api.Post(&urlPath, requestBody, &response)
	if err != nil {
		return &response, err
	}

	return &response, err
}

// DeleteObjectStorage - Submits a request to delete object storage
func (s *Service) DeleteObjectStorage(id string) error {
	urlPath := fmt.Sprintf("%s/%s/%s", s.Endpoint, ObjectStore, id)

	_, err := s.Api.Delete(&urlPath, nil, nil)
	if err != nil {
		return err
	}

	return err
}

func (s *Service) GetOrgHealthDetails() (model.OrgHealthDetails, error) {
	var response model.OrgHealthDetails

	reqUrl := fmt.Sprintf("%s/%s/%s/%s", s.Endpoint, FleetMangement, OrgHealth, Details)

	_, err := s.Api.Get(&reqUrl, nil, &response)
	if err != nil {
		return response, err
	}
	return response, nil
}

func (s *Service) GetDataplaneCounts() (model.DataplneCounts, error) {
	var response model.DataplneCounts

	reqUrl := fmt.Sprintf("%s/%s/%s/%s", s.Endpoint, FleetMangement, Dataplane, Count)

	_, err := s.Api.Get(&reqUrl, nil, &response)
	if err != nil {
		return response, err
	}
	return response, nil
}
