package index

import (
	"context"
	"encoding/json"

	"bitbucket.org/kleinnic74/photos/library"
	"bitbucket.org/kleinnic74/photos/logging"
	bolt "go.etcd.io/bbolt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type MigratableInstances interface {
	MigrateInstances(context.Context, func(int, int)) error
}

type MigratableStructure interface {
	MigrateStructure(ctx context.Context, from library.Version) (reached library.Version, reindex bool, err error)
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
			idx.Name = Name(k)
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

func (c *MigrationCoordinator) GetIndexes() (names []Name) {
	for k := range c.versions {
		names = append(names, k)
	}
	return
}

func (c *MigrationCoordinator) GetIndexStatus(name Name) (IndexStatus, bool) {
	state, found := c.versions[name]
	if !found {
		return IndexStatus{}, false
	}
	return IndexStatus{Version: state.Version}, true
}

type loggableIndexes []Name

func (l loggableIndexes) MarshalLogArray(e zapcore.ArrayEncoder) error {
	for i := range l {
		e.AppendString(string(l[i]))
	}
	return nil
}

func (c *MigrationCoordinator) Migrate(ctx context.Context, progress func(int, int)) ([]Name, error) {
	var needReindexing []Name
	logger, ctx := logging.SubFrom(ctx, "migrationCoordinator")
	logger.Info("Migrating structures")
	for name, s := range c.structures {
		currentState := c.versions[name]
		log, ctx := logging.FromWithFields(ctx, zap.String("index", string(name)), zap.Uint("currentVersion", uint(currentState.Version)))
		log.Info("Begin migration")
		nextVersion, reindex, err := s.MigrateStructure(ctx, currentState.Version)
		if err != nil {
			log.Warn("Migration failed", zap.Stringer("index", name), zap.Error(err))
		}
		if reindex {
			needReindexing = append(needReindexing, name)
		}
		currentState.Version = nextVersion
		if err := c.updateState(ctx, name, currentState); err != nil {
			log.Warn("Error while storing index status", zap.Stringer("index", name), zap.Error(err))
		}
		log.Info("migrated", zap.Uint("migratedVersion", uint(currentState.Version)))
	}
	logger.Info("Migrating instances")
	for _, i := range c.instances {
		err := i.MigrateInstances(ctx, progress)
		if err != nil {
			logger.Warn("Migration failed", zap.Error(err))
		}
	}
	logger.Info("Migration finished", zap.Array("staleIndexes", loggableIndexes(needReindexing)))
	return needReindexing, nil
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

// StructuralMigration migrates the structure of the underlying datastore
type StructuralMigration interface {
	// Apply applies the structural migration and returns a boolean indicating if reindexing is required
	Apply(ctx context.Context) (bool, error)
}

type StructuralMigrationFunc func(ctx context.Context) (bool, error)

func (m StructuralMigrationFunc) Apply(ctx context.Context) (bool, error) {
	return m(ctx)
}

// ForceReindex is a structural migration doing nothing except forcing to reindex all instances
var ForceReindex = StructuralMigrationFunc(func(context.Context) (bool, error) { return true, nil })

// StructucalMigrations are a collection of migrations to be applied to a given data store
type StructuralMigrations map[library.Version][]StructuralMigration

func NewStructuralMigrations() StructuralMigrations {
	return make(StructuralMigrations)
}

func (migrations StructuralMigrations) Register(targetVersion library.Version, m StructuralMigration) {
	migrations[targetVersion] = append(migrations[targetVersion], m)
}

func (migrations StructuralMigrations) Apply(ctx context.Context, current library.Version, target library.Version) (reindex bool, err error) {
	_, ctx = logging.SubFrom(ctx, "structural")
	for current < target {
		current++
		for _, m := range migrations[current] {
			needsReindexing, err := m.Apply(ctx)
			if err != nil {
				return false, err
			}
			reindex = reindex || needsReindexing
		}
	}
	return
}
