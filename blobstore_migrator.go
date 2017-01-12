package goblob

import (
	"errors"
	"fmt"
	"sync"

	"github.com/c0-ops/goblob/blobstore"
	"github.com/cheggaaa/pb"
	"golang.org/x/sync/errgroup"
)

// BlobstoreMigrator moves blobs from one blobstore to another
type BlobstoreMigrator interface {
	Migrate(dst blobstore.Blobstore, src blobstore.Blobstore) error
}

type blobstoreMigrator struct {
	concurrentMigrators int
	blobMigrator        BlobMigrator
}

func NewBlobstoreMigrator(concurrentMigrators int, blobMigrator BlobMigrator) BlobstoreMigrator {
	return &blobstoreMigrator{
		concurrentMigrators: concurrentMigrators,
		blobMigrator:        blobMigrator,
	}
}

func (m *blobstoreMigrator) Migrate(dst blobstore.Blobstore, src blobstore.Blobstore) error {
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

	bar := pb.StartNew(len(blobs))
	bar.Format("<.- >")

	fmt.Printf("Migrating blobs from %s to %s\n", src.Name(), dst.Name())

	blobsToMigrate := make(chan *blobstore.Blob)
	wg := sync.WaitGroup{}
	go func() {
		for _, blob := range blobs {
			wg.Add(1)
			if !dst.Exists(blob) {
				blobsToMigrate <- blob
			}
		}

		wg.Wait()
		close(blobsToMigrate)
	}()

	var g errgroup.Group
	for i := 0; i < m.concurrentMigrators; i++ {
		g.Go(func() error {
			for blob := range blobsToMigrate {
				err := m.blobMigrator.Migrate(blob)
				if err != nil {
					return err
				}
				bar.Increment()
				wg.Done()
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		bar.FinishPrint(fmt.Sprintf("Error migrating blobs from %s to %s\n", src.Name(), dst.Name()))
		return err
	}

	bar.FinishPrint(fmt.Sprintf("Done migrating blobs from %s to %s\n", src.Name(), dst.Name()))
	return nil
}
