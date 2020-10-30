package index

import (
	"context"
	"encoding/json"

	"bitbucket.org/kleinnic74/photos/library"
	"bitbucket.org/kleinnic74/photos/logging"
	"github.com/boltdb/bolt"
	"go.uber.org/zap"
)

type MigratableInstances interface {
	MigrateInstances(context.Context) error
}

type MigratableStructure interface {
	MigrateStructure(context.Context, library.Version) (library.Version, error)
}

type indexState struct {
	Name    Name            `json:"name"`
	Version library.Version `json:"version"`
}

type MigrationCoordinator struct {
	db       *bolt.DB
	versions map[Name]indexState

	instances  []MigratableInstances
	structures map[Name]MigratableStructure
}

var (
	migratablesBucket = []byte("_migratables")
)

func NewMigrationCoordinator(db *bolt.DB) (*MigrationCoordinator, error) {
	versions := make(map[Name]indexState)
	if err := db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(migratablesBucket)
		if err != nil {
			return err
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var idx indexState
			if err := json.Unmarshal(v, &idx); err != nil {
				return err
			}
			versions[Name(k)] = idx
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &MigrationCoordinator{db: db, versions: versions, structures: make(map[Name]MigratableStructure)}, nil
}

func (c *MigrationCoordinator) AddStructure(name Name, s MigratableStructure) {
	c.structures[name] = s
}

func (c *MigrationCoordinator) AddInstances(i MigratableInstances) {
	c.instances = append(c.instances, i)
}

func (c *MigrationCoordinator) Migrate(ctx context.Context) error {
	logger, ctx := logging.SubFrom(ctx, "migrationCoordinator")
	logger.Info("Migrating structures")
	for name, s := range c.structures {
		currentState := c.versions[name]
		nextVersion, err := s.MigrateStructure(ctx, currentState.Version)
		if err != nil {
			logger.Warn("Migration failed", zap.Stringer("index", name), zap.Error(err))
		}
		currentState.Version = nextVersion
		if err := c.updateState(ctx, name, currentState); err != nil {
			logger.Warn("Error while storing index status", zap.Stringer("index", name), zap.Error(err))
		}
	}
	logger.Info("Migrating instances")
	for _, i := range c.instances {
		err := i.MigrateInstances(ctx)
		if err != nil {
			logger.Warn("Migration failed", zap.Error(err))
		}
	}
	return nil
}

func (c *MigrationCoordinator) updateState(ctx context.Context, name Name, state indexState) error {
	return c.db.Update(func(tx *bolt.Tx) error {
		encoded, err := json.Marshal(&state)
		if err != nil {
			return err
		}
		b := tx.Bucket(migratablesBucket)
		return b.Put([]byte(name), encoded)
	})
}
