package model

type User struct {
	Id           string     `json:"id"`
	Email        string     `json:"email"`
	Name         string     `json:"name"`
	Status       string     `json:"status"`
	OrgRoles     []RoleMini `json:"orgRoles,omitempty"`
	ServiceRoles []RoleMini `json:"serviceRoles"`
	Tags         []string   `json:"tags"`
}
