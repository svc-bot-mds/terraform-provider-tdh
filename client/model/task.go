package model

type Task struct {
	Id          string   `json:"id"`
	TaskType    string   `json:"taskType"`
	Status      string   `json:"status"`
	DisplayName string   `json:"displayName,omitempty"`
	UiParams    UiParams `json:"uiParams"`
}

type UiParams struct {
	ResourceId   string `json:"resourceId"`
	ResourceName string `json:"resourceName"`
	ServiceType  string `json:"serviceType"`
}
