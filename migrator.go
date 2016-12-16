package goblob

import "errors"

// CloudFoundryMigrator moves blobs from Cloud Foundry to another store
type CloudFoundryMigrator struct {
}

// Migrate from a source CloudFoundry to a destination Store
func (m *CloudFoundryMigrator) Migrate(dst Store, c CloudFoundry) error {
	if c == nil {
		return errors.New("cloud foundry is empty")
	}
	store, err := c.Store()
	if err != nil {
		return err
	}
	if store == nil {
		return errors.New("src is an empty store")
	}

	if dst == nil {
		return errors.New("dst is an empty store")
	}

	files, err := store.List()
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return errors.New("the source store has no files")
	}
	_, err = dst.List()
	if err != nil {
		return err
	}

	for _, file := range files {
		/*dest := &Blob{
			Filename: file.Filename,
			Checksum: file.Checksum,
			Path: (file, c)
		}*/
		reader, err := store.Read(file)
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
