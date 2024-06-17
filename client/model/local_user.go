package model

type LocalUser struct {
	Id        string   `json:"id"`
	Username  string   `json:"username"`
	PolicyIds []string `json:"policyIds"`
}
