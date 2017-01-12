package goblob

import (
	"errors"
	"fmt"
	"sync"

	"code.cloudfoundry.org/workpool"

	"github.com/c0-ops/goblob/blobstore"
	"github.com/cheggaaa/pb"
)

var (
	buckets = []string{"cc-buildpacks", "cc-droplets", "cc-packages", "cc-resources"}
)

// BlobstoreMigrator moves blobs from one blobstore to another
type BlobstoreMigrator interface {
	Migrate(dst blobstore.Blobstore, src blobstore.Blobstore) error
}

type blobstoreMigrator struct {
	concurrentMigrators int
	pool                *workpool.WorkPool
	blobMigrator        BlobMigrator
}

func NewBlobstoreMigrator(
	pool *workpool.WorkPool,
	blobMigrator BlobMigrator,
) BlobstoreMigrator {
	return &blobstoreMigrator{
		pool:         pool,
		blobMigrator: blobMigrator,
	}
}

func (m *blobstoreMigrator) Migrate(dst blobstore.Blobstore, src blobstore.Blobstore) error {
	if src == nil {
		return errors.New("src is an empty store")
	}

	if dst == nil {
		return errors.New("dst is an empty store")
	}

	var migrateWG sync.WaitGroup
	var migrateError error
	for _, bucket := range buckets {
		iterator, err := src.NewBucketIterator(bucket)
		if err != nil {
			return fmt.Errorf("could not create bucket iterator for bucket %s: %s", bucket, err)
		}

		progressBar := pb.New(int(iterator.Count()))
		progressBar.Prefix(bucket + "\t")
		progressBar.SetMaxWidth(80)
		progressBar.Start()

		var bucketWG sync.WaitGroup

		for {
			if migrateError != nil {
				return migrateError
			}

			blob, err := iterator.Next()
			if err == blobstore.ErrIteratorDone {
				break
			}

			if err != nil {
				return err
			}

			checksum, err := src.Checksum(blob)
			if err != nil {
				return fmt.Errorf("could not checksum blob: %s", err)
			}

			blob.Checksum = checksum

			migrateWG.Add(1)
			bucketWG.Add(1)
			m.pool.Submit(func() {
				defer bucketWG.Done()
				defer migrateWG.Done()

				progressBar.Increment()

				if !dst.Exists(blob) {
					err := m.blobMigrator.Migrate(blob)
					if err != nil {
						migrateError = fmt.Errorf("error migrating %s: %s", blob.Path, err)
						return
					}
				}
			})
		}

		bucketWG.Wait()
		progressBar.Finish()
	}

	if migrateError != nil {
		return migrateError
	}

	migrateWG.Wait()

	return nil
}
