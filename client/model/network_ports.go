package model

type NetworkPorts struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Port        int64  `json:"port"`
}
