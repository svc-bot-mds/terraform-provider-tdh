package infra_connector

type DataPlaneRegionsQuery struct {
	Provider  string `schema:"provider,omitempty"`
	CPU       string `schema:"cpu,omitempty"`
	Memory    string `schema:"memory,omitempty"`
	Storage   string `schema:"storage,omitempty"`
	NodeCount string `schema:"nodeCount,omitempty"`
	OrgId     string `schema:"orgId,omitempty"`
}
