package controller

type ServiceVersionsQuery struct {
	ServiceType  string `schema:"serviceType,omitempty"`
	Provider     string `schema:"provider,omitempty"`
	Action       string `schema:"action,omitempty"`
	TemplateType string `schema:"templateType,omitempty"`
}
