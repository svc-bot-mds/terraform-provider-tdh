package model

type Extension struct {
	Id          string            `json:"id"`
	Name        string            `json:"name"`
	ServiceType string            `json:"serviceType"`
	Description string            `json:"description"`
	Version     string            `json:"version"`
	Metadata    ExtensionMetadata `json:"metadata"`
}

type ExtensionMetadata struct {
	Port string `json:"port"`
}
