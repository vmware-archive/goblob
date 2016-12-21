package goblob

import (
	"errors"
	"fmt"
	"path"

	"github.com/cheggaaa/pb"
	"golang.org/x/sync/errgroup"
)

// CloudFoundryMigrator moves blobs from Cloud Foundry to another store
type CloudFoundryMigrator struct {
	concurrentMigrators int
}

func New(concurrent int) *CloudFoundryMigrator {
	return &CloudFoundryMigrator{
		concurrentMigrators: concurrent,
	}
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

	if len(blobsToMigrate) > 0 {
		return m.migrate(dst, src, blobsToMigrate)
	}

	return nil
}

func (m *CloudFoundryMigrator) migrate(dst Store, src Store, blobs []*Blob) error {
	var g errgroup.Group

	blobsToMigrate := make(chan *Blob, len(blobs))
	for _, blob := range blobs {
		blobsToMigrate <- blob
	}
	close(blobsToMigrate)

	fmt.Println("Migrating blobs from NFS to S3")
	bar := pb.StartNew(len(blobs))
	bar.Format("<.- >")

	for i := 0; i < m.concurrentMigrators; i++ {
		g.Go(func() error {
			for blob := range blobsToMigrate {
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
					return fmt.Errorf("Checksum [%s] does not match [%s] for [%s]", checksum, blob.Checksum, path.Join(blob.Path, blob.Filename))
				}
				bar.Increment()
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		bar.FinishPrint("Error Migrating blobs from NFS to S3")
		return err
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
