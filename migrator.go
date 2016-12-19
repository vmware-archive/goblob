package goblob

import (
	"errors"
	"fmt"

	"github.com/cheggaaa/pb"
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

	var blobsToMigrate []*Blob
	for _, blob := range blobs {
		if !m.alreadyMigrated(migratedBlobs, blob) {
			blobsToMigrate = append(blobsToMigrate, blob)
		}
	}
	fmt.Println("Migrating blobs from NFS to S3")
	bar := pb.StartNew(len(blobsToMigrate))
	bar.Format("<.- >")
	for _, blob := range blobsToMigrate {
		reader, err := src.Read(blob)
		if err != nil {
			return err
		}
		err = dst.Write(blob, reader)
		if err != nil {
			return err
		}
		checksum, err := dst.Checksum(blob)
		if err != nil {
			return err
		}
		if checksum != blob.Checksum {
			return fmt.Errorf("Checksum [%s] does not match [%s]", checksum, blob.Checksum)
		}

		bar.Increment()
	}

	bar.FinishPrint("Done Migrating blobs from NFS to S3")

	return nil
}

func (m *CloudFoundryMigrator) alreadyMigrated(migratedBlobs []*Blob, blob *Blob) bool {
	for _, migratedBlob := range migratedBlobs {
		if migratedBlob.Equal(*blob) {
			return true
		}
	}
	return false
}
