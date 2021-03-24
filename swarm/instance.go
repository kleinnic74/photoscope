package swarm

type Instance struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	GitCommit string `json:"gitCommit"`
	GitRepo   string `json:"gitRepo"`
}
