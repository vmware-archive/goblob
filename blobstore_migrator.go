package goblob

import (
	"errors"
	"fmt"
	"os"
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
}

func NewBlobstoreMigrator(
	pool *workpool.WorkPool,
	blobMigrator BlobMigrator,
	exclusions []string,
) BlobstoreMigrator {
	skip := make(map[string]struct{})
	for i := range exclusions {
		skip[exclusions[i]] = struct{}{}
	}

	return &blobstoreMigrator{
		pool:         pool,
		blobMigrator: blobMigrator,
		skip:         skip,
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
	for _, bucket := range buckets {
		if _, ok := m.skip[bucket]; ok {
			continue
		}

		iterator, err := src.NewBucketIterator(bucket)
		if err != nil {
			return fmt.Errorf("could not create bucket iterator for bucket %s: %s", bucket, err)
		}

		fmt.Printf("Migrating %s ", bucket)

		var bucketWG sync.WaitGroup
		var migrateErrors []error
		for {
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

				fmt.Print(".")

				if !dst.Exists(blob) {
					err := m.blobMigrator.Migrate(blob)
					if err != nil {
						migrateErrors = append(migrateErrors, err)
						return
					}
				}
			})
		}

		bucketWG.Wait()
		fmt.Println(" done.")

		if migrateErrors != nil {
			for i := range migrateErrors {
				fmt.Fprintln(os.Stderr, migrateErrors[i])
			}
		}
	}

	migrateWG.Wait()

	return nil
}
