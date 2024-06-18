package controller

type BackupCreateRequest struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	BackupSchedule string `json:"backupSchedule"`
	BackupType     string `json:"backupType"`
}
