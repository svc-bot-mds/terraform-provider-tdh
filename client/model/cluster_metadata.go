package model

type ClusterMetaData struct {
	Id                    string                       `json:"id"`
	Name                  string                       `json:"name"`
	Provider              string                       `json:"provider"`
	ServiceType           string                       `json:"serviceType"`
	Status                string                       `json:"status"`
	VHosts                []VhostModel                 `json:"vhosts,omitempty"`
	Queues                []QueueModel                 `json:"queues,omitempty"`
	Exchanges             []ExchangeModel              `json:"exchanges,omitempty"`
	Bindings              []BindingModel               `json:"bindings,omitempty"`
	Databases             []DatabaseModel              `json:"databases,omitempty"`
	PostgresExtensionData []PostgresExtensionDataModel `json:"postgresExtensionData,omitempty"`
}

type VhostModel struct {
	Name string `json:"name"`
}

type QueueModel struct {
	Name  string `json:"name"`
	VHost string `json:"vhost"`
}

type ExchangeModel struct {
	Name  string `json:"name"`
	VHost string `json:"vhost"`
}

type BindingModel struct {
	Source          string `json:"source"`
	VHost           string `json:"vhost"`
	RoutingKey      string `json:"routingKey"`
	Destination     string `json:"destination"`
	DestinationType string `json:"destinationType"`
}

type DatabaseModel struct {
	Name     string        `json:"name"`
	Owner    string        `json:"owner"`
	Schemas  []SchemaModel `json:"schemas,omitempty"`
	Tables   []TableModel  `json:"tables,omitempty"`
	Routines []string      `json:"routines,omitempty"`
}

type SchemaModel struct {
	Name   string       `json:"name"`
	Owner  string       `json:"owner"`
	Tables []TableModel `json:"tables,omitempty"`
}

type TableModel struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Columns []string `json:"columns,omitempty"`
}

type PostgresExtensionDataModel struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}
