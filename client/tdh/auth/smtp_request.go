package auth

type SmtpRequest struct {
	Host            string `json:"host"`
	Port            string `json:"port"`
	From            string `json:"from"`
	UserName        string `json:"username"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
	TlsEnabled      string `json:"tlsEnabled"`
	Auth            string `json:"auth"`
}
