package model

type ClusterBackup struct {
	Id                string `json:"id"`
	OrgId             string `json:"orgId"`
	Name              string `json:"name"`
	GeneratedName     string `json:"generatedName"`
	ClusterId         string `json:"clusterId"`
	ClusterName       string `json:"clusterName"`
	Provider          string `json:"provider"`
	TimeStarted       string `json:"timeStarted"`
	TimeCompleted     string `json:"timeCompleted"`
	Region            string `json:"region"`
	Status            string `json:"status"`
	ClusterVersion    string `json:"clusterVersion"`
	ServiceType       string `json:"serviceType"`
	DataPlaneId       string `json:"dataPlaneId"`
	Size              string `json:"size"`
	BackupRequestId   string `json:"backupRequestId"`
	BackupTriggerType string `json:"backupTriggerType"`
	Metadata          struct {
		ClusterName        string   `json:"clusterName"`
		ClusterSize        string   `json:"clusterSize"`
		BackupLocation     string   `json:"backupLocation"`
		Databases          []string `json:"databases"`
		PostgresExtensions []struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"postgresExtensions"`
	} `json:"metadata"`
}
