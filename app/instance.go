package app

import (
	"bitbucket.org/kleinnic74/photos/consts"
	"bitbucket.org/kleinnic74/photos/swarm"
)

const (
	instanceBucket = "_instance"
)

func DefaultInstanceProperties() []swarm.PropertyDefinition {
	return []swarm.PropertyDefinition{
		swarm.WithProperty("ts", thumbCreationSpeed),
		swarm.WithPropertyValue("gc", consts.GitCommit),
		swarm.WithPropertyValue("gr", consts.GitRepo),
	}
}
