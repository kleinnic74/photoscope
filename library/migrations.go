package library

import (
	"context"
)

type InstanceMigration interface {
	// Apply will apply this migration to the given photo, passing a ReaderFunc to optionally
	// access the photo content
	Apply(context.Context, Photo, ReaderFunc) (Photo, error)
}

type InstanceFunc func(context.Context, Photo, ReaderFunc) (Photo, error)

func (m InstanceFunc) Apply(ctx context.Context, p Photo, c ReaderFunc) (Photo, error) {
	return m(ctx, p, c)
}

type InstanceMigrations map[Version][]InstanceMigration

func NewInstanceMigrations() InstanceMigrations {
	return make(InstanceMigrations)
}

func (migrations InstanceMigrations) Register(toTargetSchema Version, m InstanceMigration) {
	all := append(migrations[toTargetSchema], m)
	migrations[toTargetSchema] = all
}

func (migrations InstanceMigrations) Apply(ctx context.Context, photo Photo, content ReaderFunc) (result Photo, err error) {
	result = photo
	for result.schema < currentSchema {
		nextSchema := photo.schema + 1
		m := migrations[nextSchema]
		for _, migration := range m {
			result, err = migration.Apply(ctx, result, content)
			if err != nil {
				return
			}
		}
		result.schema = nextSchema
	}
	return
}

type StructuralMigration interface {
	Apply() error
}

type StructuralMigrationFunc func() error

func (m StructuralMigrationFunc) Apply() error {
	return m()
}

type StructuralMigrations map[Version][]StructuralMigration

func NewStructuralMigrations() StructuralMigrations {
	return make(StructuralMigrations)
}

func (migrations StructuralMigrations) Register(targetVersion Version, m StructuralMigration) {
	migrations[targetVersion] = append(migrations[targetVersion], m)
}

func (migrations StructuralMigrations) Apply(current Version, target Version) error {
	for current < target {
		for _, m := range migrations[current] {
			if err := m.Apply(); err != nil {
				return err
			}
		}
		current++
	}
	return nil
}
