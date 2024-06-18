package controller

type ServiceVersionsQuery struct {
	ServiceType  string `schema:"serviceType"`
	Provider     string `schema:"provider"`
	Action       string `schema:"action"`
	TemplateType string `schema:"templateType"`
}
