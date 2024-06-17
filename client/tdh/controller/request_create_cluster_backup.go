package controller

type BackupCreateRequest struct {
	ClusterId      string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	BackupSchedule string `json:"backupSchedule"`
	BackupType     string `json:"backupType"`
	ServiceType    string `json:"serviceType"`
}
