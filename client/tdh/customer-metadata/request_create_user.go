package customer_metadata

type CreateUserRequest struct {
	AccountType   string         `json:"accountType"`
	Usernames     []string       `json:"usernames"` // List of emails by which to invite/add the users.
	PolicyIds     []string       `json:"policyIds"`
	ServiceRoles  []RolesRequest `json:"serviceRoles"`
	Tags          []string       `json:"tags"`
	Organizations []string       `json:"orgs,omitempty"`
}

type RolesRequest struct {
	RoleId string `json:"roleId,omitempty"`
}
