package model

type Certificate struct {
	Id          string      `json:"id"`
	DomainName  string      `json:"domainName"`
	Name        string      `json:"name"`
	Provider    string      `json:"provider"`
	ExpiryTime  string      `json:"expirationTime"`
	CreatedBy   string      `json:"createdBy,omitempty"`
	OrgId       string      `json:"orgId"`
	Status      string      `json:"status"`
	Deployments interface{} `json:"resources,omitempty"`
	Tags        []string    `json:"tags,omitempty"`
}
