package model

type DataPlane struct {
	Id            string   `json:"id"`
	Provider      string   `json:"provider"`
	Region        string   `json:"region,omitempty"`
	Name          string   `json:"name"`
	DataplaneName string   `json:"dataplaneName"`
	Version       string   `json:"version"`
	Tags          []string `json:"tags"`
	Status        string   `json:"status"`
	NodePoolType  string   `json:"nodePoolType"`
	Account       struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	} `json:"account"`
	Upgradeable          bool   `json:"upgradeable"`
	AutoUpgrade          bool   `json:"autoUpgrade"`
	DataPlaneReleaseID   string `json:"dataPlaneReleaseID"`
	DataPlaneReleaseName string `json:"dataPlaneReleaseName"`
	Shared               bool   `json:"shared"`
	Created              string `json:"created"`
	Modified             string `json:"modified"`
	Certificate          struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	} `json:"certificate"`
	ManagedDns              bool     `json:"managedDns"`
	DefaultPolicyName       string   `json:"defaultPolicyName"`
	StoragePolicies         []string `json:"storagePolicies"`
	BackupStoragePolicy     string   `json:"backupStoragePolicy"`
	Services                []string `json:"services"`
	LoggingUri              string   `json:"logging_uri"`
	InfraResourceType       string   `json:"infraResourceType"`
	DataPlaneOnControlPlane bool     `json:"dataPlaneOnControlPlane"`
	OrgId                   string   `json:"orgId,omitempty"`
}

type HelmVersions struct {
	Id        string `json:"id"`
	Name      string `json:"releaseName"`
	IsEnabled bool   `json:"isEnabled"`
}

type EligibleSharedDataPlane struct {
	Id                  string   `json:"id"`
	Provider            string   `json:"provider"`
	DataplaneName       string   `json:"dataplaneName"`
	StoragePolicies     []string `json:"storagePolicies"`
	BackupStoragePolicy string   `json:"backupStoragePolicy"`
}

type TKC struct {
	Name        string `json:"clusterName"`
	IsAvailable bool   `json:"isAvailable"`
	IsCpPresent bool   `json:"isCPPresent"`
}
