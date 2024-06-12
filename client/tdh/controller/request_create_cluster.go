package controller

type ClusterCreateRequest struct {
	Name              string          `json:"name"`
	ServiceType       string          `json:"serviceType"`
	Provider          string          `json:"provider"`
	InstanceSize      string          `json:"instanceSize"`
	Region            string          `json:"region"`
	Dedicated         bool            `json:"dedicated"`
	Shared            bool            `json:"shared,omitempty"`
	Tags              []string        `json:"tags,omitempty"`
	NetworkPolicyIds  []string        `json:"networkPolicyIds,omitempty"`
	DataPlaneId       string          `json:"dataPlaneId,omitempty"`
	Version           string          `json:"version"`
	StoragePolicyName string          `json:"storagePolicyName"`
	ClusterMetadata   ClusterMetadata `json:"clusterMetadata"`
}

type ClusterMetadata struct {
	Username      string   `json:"username"`
	Password      string   `json:"password"`
	Database      string   `json:"database"`
	RestoreFrom   string   `json:"restore_from"`
	ObjectStoreId string   `json:"ObjectStoreId"`
	Extensions    []string `json:"extensions"`
}
