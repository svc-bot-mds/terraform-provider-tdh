package infra_connector

type StorageClassesQuery struct {
	AccountId   string `schema:"accountId,omitempty"`
	ClusterName string `schema:"clusterName,omitempty"`
}
