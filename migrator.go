package goblob

import "errors"

// CloudFoundryMigrator moves blobs from Cloud Foundry to another store
type CloudFoundryMigrator struct {
}

// Migrate from a source CloudFoundry to a destination Store
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

	for _, file := range files {
		reader, err := src.Read(file)
		if err != nil {
			return err
		}
		err = dst.Write(file, reader)
		if err != nil {
			return err
		}
	}

	return nil
}
