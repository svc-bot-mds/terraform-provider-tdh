package customer_metadata

type SvcAccountUpdateRequest struct {
	Tags      []string `json:"tags"`
	PolicyIds []string `json:"policyIds"`
}
