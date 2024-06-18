package controller

type RestoreClusterBackupRequest struct {
	Name              string          `json:"name"`
	Tags              []string        `json:"tags"`
	Provider          string          `json:"provider"`
	Version           string          `json:"version"`
	NetworkPolicyIds  []string        `json:"networkPolicyIds"`
	Region            string          `json:"region"`
	InstanceSize      string          `json:"instanceSize"`
	ServiceType       string          `json:"serviceType"`
	Dedicated         bool            `json:"dedicated"`
	Shared            bool            `json:"shared"`
	StoragePolicyName string          `json:"storagePolicyName"`
	ClusterMetadata   ClusterMetadata `json:"clusterMetadata"`
}
