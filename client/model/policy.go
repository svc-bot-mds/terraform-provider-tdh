package model

// Policy base model for TDH Policy
type Policy struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Description     string            `json:"description"`
	ServiceType     string            `json:"serviceType"`
	Updating        bool              `json:"updating"`
	ResourceIds     []string          `json:"resourceIds,omitempty"`
	PermissionsSpec []PermissionsSpec `json:"permissionsSpec,omitempty"`
	NetworkSpec     []*NetworkSpec    `json:"networkSpecs,omitempty"`
}
type PermissionsSpec struct {
	Resource    string         `json:"resource"`
	Permissions []*Permissions `json:"permissions"`
	Role        string         `json:"role"`
}

type NetworkSpec struct {
	CIDR           string   `json:"cidr"`
	NetworkPortIds []string `json:"networkPortIds"`
}

type Permissions struct {
	Name         string `json:"name"`
	PermissionId string `json:"permissionId"`
}
