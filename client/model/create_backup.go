package model

type CreateBackup struct {
	Id             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	BackupSchedule string `json:"backupSchedule"`
	BackupType     string `json:"backupType"`
}
