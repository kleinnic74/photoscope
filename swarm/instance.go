package swarm

import (
	"context"
	"fmt"
	"os"
	"strings"
)

const (
	idKey = "id"
)

type InstanceID string

func (i InstanceID) String() string {
	return string(i)
}

type Instance struct {
	ID         InstanceID        `json:"id"`
	Name       string            `json:"name"`
	Properties map[string]string `json:"properties,omitempty"`
}

type PropertyProvider func(context.Context) string
type PropertyDefinition func() (string, PropertyProvider)

func WithProperty(name string, f PropertyProvider) PropertyDefinition {
	return func() (string, PropertyProvider) {
		return name, f
	}
}

func WithPropertyValue(name string, value string) PropertyDefinition {
	return func() (string, PropertyProvider) {
		return name, func(context.Context) string { return value }
	}
}

func NewInstance(ctx context.Context, id InstanceID, p ...PropertyDefinition) (*Instance, error) {
	hostnameFQ, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	hostname := strings.Split(hostnameFQ, ".")[0]
	i := &Instance{
		ID:         id,
		Name:       fmt.Sprintf("Photoscope on %s", hostname),
		Properties: make(map[string]string),
	}
	i.Properties[idKey] = i.ID.String()
	for _, pd := range p {
		name, f := pd()
		i.Properties[name] = f(ctx)
	}
	return i, err
}
