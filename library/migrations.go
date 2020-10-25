package library

import (
	"context"
	"io"

	"bitbucket.org/kleinnic74/photos/consts"
	"bitbucket.org/kleinnic74/photos/domain"
	"bitbucket.org/kleinnic74/photos/logging"
	"go.uber.org/zap"
)

type Migration interface {
	// Apply will apply this migration to the given photo, passing a ReaderFunc to optionally
	// access the photo content
	Apply(context.Context, Photo, ReaderFunc) (Photo, error)
}

type MigrationFunc func(context.Context, Photo, ReaderFunc) (Photo, error)

func (m MigrationFunc) Apply(ctx context.Context, p Photo, c ReaderFunc) (Photo, error) {
	return m(ctx, p, c)
}

var migrations map[uint][]Migration

func init() {
	migrations = make(map[uint][]Migration)
}

func RegisterMigration(toTargetSchema uint, m Migration) {
	all := append(migrations[toTargetSchema], m)
	migrations[toTargetSchema] = all
}

func ApplyMigrations(ctx context.Context, photo Photo, content ReaderFunc) (result Photo, changed bool, err error) {
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
	changed = result != photo
	return
}

func (lib *BasicPhotoLibrary) UpgradeDBStructures(ctx context.Context) error {
	logger, ctx := logging.SubFrom(ctx, "upgradeDB")
	photos, err := lib.FindAll(ctx, consts.Ascending)
	if err != nil {
		return err
	}
	var count int
	for _, p := range photos {
		updated, changed, err := ApplyMigrations(ctx, *p, func() (io.ReadCloser, error) {
			return lib.openPhoto(p.Path)
		})
		if err != nil {
			logger.Warn("Migration failed", zap.String("photo", string(p.ID)))
			continue
		}
		if changed {
			logger.Info("Photo migrated", zap.String("photo", string(p.ID)))
			lib.db.Update(&updated)
			count++
		}
	}
	if count > 0 {
		logger.Info("Fixed photos in DB", zap.Int("count", count))
	}
	return nil

}

func migratePath(ctx context.Context, p Photo, _ ReaderFunc) (Photo, error) {
	logger, ctx := logging.SubFrom(ctx, "migratePath")
	if !isPathConversionNeeded(p.Path) {
		return p, nil
	}
	oldPath := p.Path
	p.Path = convertPath(oldPath)
	logger.Info("Fixed photo path", zap.String("photo", string(p.ID)), zap.String("path", p.Path), zap.String("oldpath", oldPath))
	return p, nil
}

func migrateHash(ctx context.Context, p Photo, in ReaderFunc) (Photo, error) {
	if !p.HasHash() {
		content, err := in()
		if err != nil {
			return p, err
		}
		defer content.Close()
		h, err := ComputeHash(content)
		if err != nil {
			return p, err
		}
		p.Hash = h
	}
	return p, nil
}

func addOrientation(ctx context.Context, p Photo, in ReaderFunc) (Photo, error) {
	if p.Orientation == domain.UnknownOrientation {
		content, err := in()
		if err != nil {
			return p, err
		}
		defer content.Close()
		var meta domain.MediaMetaData
		if err := p.Format.DecodeMetaData(content, &meta); err != nil {
			return p, err
		}
		p.Orientation = meta.Orientation
	}
	return p, nil
}

func init() {
	RegisterMigration(1, MigrationFunc(migratePath))
	RegisterMigration(1, MigrationFunc(migrateHash))
	RegisterMigration(1, MigrationFunc(addOrientation))
}
