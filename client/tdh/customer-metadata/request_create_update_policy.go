package customer_metadata

type CreateUpdatePolicyRequest struct {
	Name            string                  `json:"name"`
	Description     string                  `json:"description"`
	ServiceType     string                  `json:"serviceType"`
	PermissionsSpec []PermissionSpecRequest `json:"permissionsSpec,omitempty"`
	NetworkSpecs    []NetworkSpec           `json:"networkSpecs,omitempty"`
}

type PermissionSpecRequest struct {
	Resource    string   `json:"resource"`
	Permissions []string `json:"permissions"`
	Role        string   `json:"role"`
}

type NetworkSpec struct {
	Cidr           string   `json:"cidr"`
	NetworkPortIds []string `json:"networkPortIds"`
}
