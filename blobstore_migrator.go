package goblob

import (
	"errors"
	"fmt"
	"sync"

	"code.cloudfoundry.org/workpool"

	"github.com/pivotalservices/goblob/blobstore"
)

var (
	buckets = []string{"cc-buildpacks", "cc-droplets", "cc-packages", "cc-resources"}
)

// BlobstoreMigrator moves blobs from one blobstore to another
type BlobstoreMigrator interface {
	Migrate(dst blobstore.Blobstore, src blobstore.Blobstore) error
}

type blobstoreMigrator struct {
	pool         *workpool.WorkPool
	blobMigrator BlobMigrator
	skip         map[string]struct{}
	watcher      BlobstoreMigrationWatcher
}

func NewBlobstoreMigrator(
	pool *workpool.WorkPool,
	blobMigrator BlobMigrator,
	exclusions []string,
	watcher BlobstoreMigrationWatcher,
) BlobstoreMigrator {
	skip := make(map[string]struct{})
	for i := range exclusions {
		skip[exclusions[i]] = struct{}{}
	}

	return &blobstoreMigrator{
		pool:         pool,
		blobMigrator: blobMigrator,
		skip:         skip,
		watcher:      watcher,
	}
}

func (m *blobstoreMigrator) Migrate(dst blobstore.Blobstore, src blobstore.Blobstore) error {
	if src == nil {
		return errors.New("src is an empty store")
	}

	if dst == nil {
		return errors.New("dst is an empty store")
	}

	m.watcher.MigrationDidStart(dst, src)

	migrateWG := &sync.WaitGroup{}
	for _, bucket := range buckets {
		if _, ok := m.skip[bucket]; ok {
			continue
		}

		iterator, err := src.NewBucketIterator(bucket)
		if err != nil {
			return fmt.Errorf("could not create bucket iterator for bucket %s: %s", bucket, err)
		}

		m.watcher.MigrateBucketDidStart(bucket)

		bucketWG := &sync.WaitGroup{}
		for {
			blob, err := iterator.Next()
			if err == blobstore.ErrIteratorDone {
				break
			}

			if err != nil {
				return err
			}

			migrateWG.Add(1)
			bucketWG.Add(1)
			m.pool.Submit(func() {
				defer bucketWG.Done()
				defer migrateWG.Done()

				checksum, err := src.Checksum(blob)
				if err != nil {
					checksumErr := fmt.Errorf("could not checksum blob: %s", err)
					m.watcher.MigrateBlobDidFailWithError(checksumErr)
					return
				}

				blob.Checksum = checksum

				if dst.Exists(blob) {
					m.watcher.MigrateBlobAlreadyFinished()
					return
				}

				err = m.blobMigrator.Migrate(blob)
				if err != nil {
					m.watcher.MigrateBlobDidFailWithError(err)
					return
				}

				m.watcher.MigrateBlobDidFinish()
			})
		}

		bucketWG.Wait()
		m.watcher.MigrateBucketDidFinish()
	}

	migrateWG.Wait()
	m.watcher.MigrationDidFinish()

	return nil
}
