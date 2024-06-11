package model

type InstanceTypeList struct {
	InstanceTypes []InstanceType `json:"instanceTypes"`
}

// InstanceType -
type InstanceType struct {
	ID              string               `json:"id,omitempty"`
	ServiceType     string               `json:"serviceType"`
	InstanceSize    string               `json:"instanceSize"`
	SizeDescription string               `json:"instanceSizeDescription"`
	CPU             string               `json:"cpu"`
	Memory          string               `json:"memory"`
	Storage         string               `json:"storage"`
	Metadata        InstanceTypeMetadata `json:"metadata,omitempty"`
}

type InstanceTypeMetadata struct {
	MaxConnections string `json:"max_connections,omitempty"`
	Nodes          string `json:"nodes,omitempty"`
}
