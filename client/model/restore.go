package model

type ClusterRestore struct {
	Id                 string `json:"id"`
	Name               string `json:"name"`
	DataPlaneId        string `json:"dataPlaneId"`
	ServiceType        string `json:"serviceType"`
	BackupId           string `json:"backupId"`
	BackupName         string `json:"backupName"`
	TargetInstance     string `json:"targetInstance"`
	TargetInstanceName string `json:"targetInstanceName"`
}
