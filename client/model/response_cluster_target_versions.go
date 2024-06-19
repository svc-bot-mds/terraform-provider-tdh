package model

type ClusterTargetVersionsResponse struct {
	ClusterId      string   `json:"clusterId"`
	Version        string   `json:"version"`
	TargetVersions []string `json:"targetVersions"`
}
