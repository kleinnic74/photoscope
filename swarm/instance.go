package swarm

type Instance struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Properties map[string]string `json:"properties,omitempty"`
}
