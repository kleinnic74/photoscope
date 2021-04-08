package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"bitbucket.org/kleinnic74/photos/swarm"
	"github.com/google/uuid"
	bolt "go.etcd.io/bbolt"
)

type InstanceStore struct {
	db *bolt.DB
	I  *swarm.Instance
}

const (
	instanceBucket = "_instance"
	idKey          = "id"
)

type PropertyProvider func() string
type PropertyDefinition func() (string, PropertyProvider)

func NewInstance(db *bolt.DB, p ...PropertyDefinition) (*InstanceStore, error) {
	hostnameFQ, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	hostname := strings.Split(hostnameFQ, ".")[0]
	store := &InstanceStore{
		db: db,
		I: &swarm.Instance{
			Properties: make(map[string]string),
		},
	}
	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(instanceBucket))
		if err != nil {
			return err
		}
		if v := b.Get([]byte(idKey)); v != nil {
			json.Unmarshal(v, store.I)
		}
		if store.I.ID == "" {
			i, err := uuid.NewRandom()
			if err != nil {
				return err
			}
			store.I.ID = i.String()
		}
		for _, pd := range p {
			name, f := pd()
			store.I.Properties[name] = f()
		}
		store.I.Name = fmt.Sprintf("Photoscope on %s", hostname)
		v, _ := json.Marshal(store.I)
		return b.Put([]byte(idKey), v)
	})
	return store, err
}
