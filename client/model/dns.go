package model

type Dns struct {
	Id         string    `json:"id"`
	Domain     string    `json:"domain"`
	Name       string    `json:"name"`
	Provider   string    `json:"provider"`
	ServerList []Servers `json:"servers"`
}

type Servers struct {
	Host       string `json:host`
	Port       int64  `json:port`
	Protocol   string `json: protocol`
	ServerType string `json: serverType`
}
