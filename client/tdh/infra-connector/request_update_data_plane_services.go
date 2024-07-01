package infra_connector

type DataPlaneUpdateServicesRequest struct {
	DataPlaneId string   `json:"dataPlaneId"`
	Services    []string `json:"services"`
}
