package customer_metadata

type CreateLocalUserRequest struct {
	Usernames       []string `json:"usernames"` // List of usernames by which to invite/add the local users.
	PolicyIds       []string `json:"policyIds"`
	Password        string   `json:"password"`
	ConfirmPassword string   `json:"confirmPassword"`
}
