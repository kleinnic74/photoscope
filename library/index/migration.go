package index

type Migration interface {
	Apply() error
}

type MigrationFunc func() error

func (m MigrationFunc) Apply() error {
	return m()
}

type MigrationSpec struct {
	Target    Version
	Migration Migration
}
