package goblob

import (
	"errors"
	"fmt"

	"github.com/c0-ops/goblob/validation"
)

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

	blobs, err := src.List()
	if err != nil {
		return err
	}

	if len(blobs) == 0 {
		return errors.New("the source store has no files")
	}

	migratedBlobs, err := dst.List()
	if err != nil {
		return err
	}
	for _, blob := range blobs {
		if !m.alreadyMigrated(migratedBlobs, blob) {
			reader, err := src.Read(blob)
			if err != nil {
				return err
			}
			err = dst.Write(blob, reader)
			if err != nil {
				return err
			}
			reader, err = dst.Read(blob)
			if err != nil {
				return err
			}
			checksum, err := validation.ChecksumReader(reader)
			if err != nil {
				return err
			}
			if checksum != blob.Checksum {
				return fmt.Errorf("Checksum [%s] does not match [%s]", checksum, blob.Checksum)
			}
		}
	}

	return nil
}

func (m *CloudFoundryMigrator) alreadyMigrated(migratedBlobs []Blob, blob Blob) bool {
	for _, migratedBlob := range migratedBlobs {
		if migratedBlob.Equal(blob) {
			return true
		}
	}
	return false
}
