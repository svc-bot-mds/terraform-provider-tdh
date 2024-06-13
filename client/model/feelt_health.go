package model

type OrgHealthDetails struct {
	TotalHealthyOrgsCount   int64 `json:"totalHealthyOrgsCount"`
	TotalOrgCount           int64 `json:"totalOrgCount"`
	TotalUnhealthyOrgsCount int64 `json:"totalUnhealthyOrgsCount"`
}

type DataplneCounts struct {
	SharedDataplanes    int64 `json:"sharedDataplanes"`
	DedicatedDataplanes int64 `json:"dedicatedDataplanes"`
	HealthyDataplanes   int64 `json:"healthyDataplanes"`
	UnhealthyDataplanes int64 `json:"unhealthyDataplanes"`
	TotalDataplanes     int64 `json:"totalDataplanes"`
}

type ClusterCountByService struct {
	Count       int64  `json:"count"`
	ServiceType string `json:"serviceType"`
}

type ResourceByService struct {
	DataPlaneName string `json:"dataPlaneName"`
	ServiceType   string `json:"serviceType"`
	Cpu           string `json:"cpu"`
	Memory        string `json:"memory"`
	Storage       string `json:"storage"`
}

type SreCustomerInfo struct {
	Name          string `json:"name"`
	OrgName       string `json:"orgName"`
	ClusterStatus struct {
		Critical int64 `json:"critical"`
		Warning  int64 `json:"warning"`
		Healthy  int64 `json:"healthy"`
	} `json:"clusterStatus"`
	ClusterCounts       int64 `json:"clusterCounts"`
	CustomerClusterInfo []struct {
		ClusterId    string `json:"clusterId"`
		ClusterName  string `json:"clusterName"`
		InstanceSize string `json:"instanceSize"`
		Status       string `json:"status"`
		ServiceType  string `json:"serviceType"`
	} `json:"mdsCustomerClusterInfo,omitempty"`
	CustomerCumulativeStatus string   `json:"mdsCustomerCumulativeStatus"`
	Services                 []string `json:"services"`
	SreOrg                   bool     `json:"sreOrg"`
}
