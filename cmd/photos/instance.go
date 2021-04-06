package main

import (
	"encoding/json"

	"bitbucket.org/kleinnic74/photos/consts"
	"github.com/google/uuid"
	bolt "go.etcd.io/bbolt"
)

type Instance struct {
	ID        string `json:"id"`
	GitCommit string `json:"gitCommit"`
	GitRepo   string `json:"gitRepo"`
}

const (
	instanceBucket = "_instance"
	idKey          = "id"
)

func NewInstance(db *bolt.DB) (*Instance, error) {
	id := &Instance{}
	err := db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(instanceBucket))
		if err != nil {
			return err
		}
		if v := b.Get([]byte(idKey)); v != nil {
			json.Unmarshal(v, id)
		}
		id.GitCommit = consts.GitCommit
		id.GitRepo = consts.GitRepo
		if id.ID == "" {
			i, err := uuid.NewRandom()
			if err != nil {
				return err
			}
			id.ID = i.String()
		}
		v, _ := json.Marshal(id)
		return b.Put([]byte(idKey), v)
	})
	return id, err
}
