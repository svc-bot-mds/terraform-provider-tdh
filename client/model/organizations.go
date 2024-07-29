package model

type OrgModel struct {
	Name           string `json:"name"`
	OrgId          string `json:"orgId"`
	ShortOrgId     string `json:"shortOrgId"`
	OrgName        string `json:"orgName"`
	Created        string `json:"created"`
	Modified       string `json:"modified"`
	SkipOrgUpdates bool   `json:"skipOrgUpdates"`
	Status         string `json:"status"`
	SreOrg         bool   `json:"sreOrg"`
}
