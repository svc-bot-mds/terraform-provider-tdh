package model

type DataPlane struct {
	Id                   string               `json:"id"`
	Provider             string               `json:"provider"`
	Name                 string               `json:"name"`
	Region               string               `json:"region"`
	K8SVersion           string               `json:"version"`
	Certificate          DataPlaneCertificate `json:"certificate"`
	DataPlaneReleaseName string               `json:"dataPlaneReleaseName"`
	Status               string               `json:"status"`
	TshirtSize           string               `json:"nodePoolType"`
}

type DataPlaneCertificate struct {
	DomainName string `json:"domainName"`
	Name       string `json:"name"`
}
