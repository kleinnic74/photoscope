package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"bitbucket.org/kleinnic74/photos/consts"
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

func NewInstance(db *bolt.DB) (*InstanceStore, error) {
	hostnameFQ, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	hostname := strings.Split(hostnameFQ, ".")[0]
	store := &InstanceStore{
		db: db,
		I:  &swarm.Instance{},
	}
	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(instanceBucket))
		if err != nil {
			return err
		}
		if v := b.Get([]byte(idKey)); v != nil {
			json.Unmarshal(v, store.I)
		}
		store.I.GitCommit = consts.GitCommit
		store.I.GitRepo = consts.GitRepo
		store.I.Name = fmt.Sprintf("Photoscope on %s", hostname)
		if store.I.ID == "" {
			i, err := uuid.NewRandom()
			if err != nil {
				return err
			}
			store.I.ID = i.String()
		}
		v, _ := json.Marshal(store.I)
		return b.Put([]byte(idKey), v)
	})
	return store, err
}
