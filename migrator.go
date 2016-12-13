package goblob

import "errors"

type CloudFoundryMigrator struct {
}

func (m *CloudFoundryMigrator) Migrate(dst Store, src Store) error {
	if src == nil {
		return errors.New("src is an empty store")
	}

	if dst == nil {
		return errors.New("dst is an empty store")
	}

	files, err := src.List()
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return errors.New("the source store has no files")
	}

	return nil
}
