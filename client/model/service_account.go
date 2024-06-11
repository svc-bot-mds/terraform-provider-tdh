package model

type ServiceAccount struct {
	Id     string   `json:"id"`
	Name   string   `json:"name"`
	Status string   `json:"status,omitempty"`
	Tags   []string `json:"tags"`
}

type ServiceAccountCreate struct {
	OAuthCredentials []*ServiceAccountAuthCredentials `json:"oauthCredentials,omitempty"`
}
type ServiceAccountAuthCredentials struct {
	UserName   string                     `json:"username,omitempty"`
	Credential *ServiceAccountCredentials `json:"credential,omitempty"`
}
type ServiceAccountCredentials struct {
	ClientId     string `json:"clientId,omitempty"`
	ClientSecret string `json:"clientSecret,omitempty"`
	GrantType    string `json:"grantType,omitempty"`
	OrgId        string `json:"orgId,omitempty"`
}

type ServiceAccountOauthApp struct {
	AppId       string   `json:"appId"`
	AppType     string   `json:"appType"`
	Created     string   `json:"created"`
	CreatedBy   string   `json:"createdBy"`
	Description string   `json:"description"`
	Modified    string   `json:"modified"`
	ModifiedBy  string   `json:"modifiedBy"`
	TTLSpec     *TTLSpec `json:"ttlSpec"`
}

type TTLSpec struct {
	Description string `json:"description,omitempty"`
	TimeUnit    string `json:"timeUnit"`
	TTL         int64  `json:"ttl"`
}
