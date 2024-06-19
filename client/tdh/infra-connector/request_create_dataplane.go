package infra_connector

type DataPlaneCreateRequest struct {
	AccountId             string   `json:"accountId"`
	DataplaneName         string   `json:"dataplaneName"`
	K8sClusterName        string   `json:"k8sClusterName"`
	DataplaneType         string   `json:"dataplaneType"`
	CertificateId         string   `json:"certificateId"`
	StorageClasses        []string `json:"storageClasses"`
	BackupStorageClass    string   `json:"backupStorageClass"`
	ManagedDns            bool     `json:"managedDns"`
	DataPlaneReleaseId    string   `json:"dataPlaneReleaseId"`
	Shared                bool     `json:"shared"`
	OrgId                 string   `json:"orgId"`
	Tags                  []string `json:"tags"`
	AutoUpgrade           bool     `json:"autoUpgrade"`
	Services              []string `json:"services"`
	CpBootstrappedCluster bool     `json:"cpBootstrappedCluster"`
	ConfigureCoreDns      bool     `json:"configureCoreDns"`
	DnsConfigId           string   `json:"dnsConfigId"`
}

type DataPlaneUpdateRequest struct {
	DataplaneName string   `json:"dataplaneName"`
	Tags          []string `json:"tags"`
	AutoUpgrade   bool     `json:"autoUpgrade"`
}
