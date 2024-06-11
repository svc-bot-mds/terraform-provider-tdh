package model

type ClusterMetaData struct {
	Id          string           `json:"id"`
	Name        string           `json:"name"`
	Provider    string           `json:"provider"`
	ServiceType string           `json:"serviceType"`
	Status      string           `json:"status"`
	VHosts      []Vhosts         `json:"vhosts,omitempty"`
	Queues      []QueuesModel    `json:"queues,omitempty"`
	Exchanges   []ExchangesModel `json:"exchanges,omitempty"`
	Bindings    []BindingsModel  `json:"bindings,omitempty"`
}

type Vhosts struct {
	Name string `json:"name"`
}

type QueuesModel struct {
	Name  string `json:"name"`
	VHost string `json:"vhost"`
}

type ExchangesModel struct {
	Name  string `json:"name"`
	VHost string `json:"vhost"`
}

type BindingsModel struct {
	Source          string `json:"source"`
	VHost           string `json:"vhost"`
	RoutingKey      string `json:"routingKey"`
	Destination     string `json:"destination"`
	DestinationType string `json:"destinationType"`
}
