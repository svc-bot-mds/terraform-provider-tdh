package model

type StorageClass struct {
	StorageClassName string `json:"storageClassName"`
	Provisioner      string `json:"provisioner"`
}
