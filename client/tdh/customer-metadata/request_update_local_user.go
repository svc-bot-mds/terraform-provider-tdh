package customer_metadata

type LocalUserUpdateRequest struct {
	PolicyIds          []string `json:"policyIds"`
	CurrentPassword    string   `json:"currentPassword"`
	NewPassword        string   `json:"newPassword"`
	ConfirmNewPassword string   `json:"confirmNewPassword"`
}
