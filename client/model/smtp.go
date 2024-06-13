package model

type Smtp struct {
	Auth            string `json:"auth"`
	Tls             string `json:"tlsEnabled"`
	FromEmail       string `json:"from"`
	UserName        string `json:"username"`
	Host            string `json:"host"`
	Port            string `json:"port"`
	Password        string `json:"password,omitempty"`
	ConfirmPassword string `json:"confirmPassword,omitempty"`
}
