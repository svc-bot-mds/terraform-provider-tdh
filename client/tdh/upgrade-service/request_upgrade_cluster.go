package upgrade_service

// UpdateClusterVersionRequest represents the request structure for updating the cluster version
type UpdateClusterVersionRequest struct {
	Id            string                              `json:"id"`
	RequestType   string                              `json:"requestType"`
	TargetVersion string                              `json:"targetVersion"`
	Metadata      UpdateClusterVersionRequestMetadata `json:"metadata"`
}

// UpdateClusterVersionRequestMetadata represents the metadata for the version update request
type UpdateClusterVersionRequestMetadata struct {
	OmitBackup string `json:"omitBackup"`
}
