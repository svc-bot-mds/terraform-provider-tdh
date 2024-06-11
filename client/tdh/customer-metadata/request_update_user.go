package customer_metadata

type UserUpdateRequest struct {
	Tags         []string       `json:"tags"`
	PolicyIds    []string       `json:"policyIds"`
	ServiceRoles []RolesRequest `json:"serviceRoles,omitempty"`
}
