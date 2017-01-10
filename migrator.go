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
	blobMigrator        BlobMigrator
}

type StatusBar interface {
	Increment() int
	FinishPrint(str string)
}

type BlobMigrator interface {
	MigrateSingleBlob(blob *Blob) error
	Init(dst Store, src Store, bar StatusBar)
	Finish(msg string)
	SingleBlobError(blob *Blob, err error) error
}

type BlobMigrate struct {
	bar StatusBar
	dst Store
	src Store
}

func (s *BlobMigrate) Init(dst Store, src Store, bar StatusBar) {
	s.dst = dst
	s.src = src
	s.bar = bar
}

func (s *BlobMigrate) SingleBlobError(blob *Blob, err error) error {
	return fmt.Errorf("error at %s: %s", path.Join(blob.Path, blob.Filename), err.Error())
}

func (s *BlobMigrate) Finish(msg string) {
	s.bar.FinishPrint(msg)
}

func (s *BlobMigrate) MigrateSingleBlob(blob *Blob) error {
	reader, err := s.src.Read(blob)
	if err != nil {
		return s.SingleBlobError(blob, err)
	}
	defer reader.Close()

	err = s.dst.Write(blob, reader)
	if err != nil {
		return s.SingleBlobError(blob, err)
	}
	checksum, err := s.dst.Checksum(blob)
	if err != nil {
		return s.SingleBlobError(blob, err)
	}
	if checksum != blob.Checksum {
		err = fmt.Errorf("Checksum [%s] does not match [%s]", checksum, blob.Checksum)
		return s.SingleBlobError(blob, err)
	}
	s.bar.Increment()
	return nil
}

func New(concurrent int) *CloudFoundryMigrator {
	return &CloudFoundryMigrator{
		concurrentMigrators: concurrent,
		blobMigrator:        new(BlobMigrate),
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

	bar := pb.StartNew(len(blobs))
	bar.Format("<.- >")
	m.blobMigrator.Init(dst, src, bar)

	fmt.Printf("Migrating blobs from %s to %s\n", src.Name(), dst.Name())

	blobsToMigrate := make(chan *Blob)
	go func() {
		for _, blob := range blobs {
			if !dst.Exists(blob) {
				blobsToMigrate <- blob
			}
		}

		close(blobsToMigrate)
	}()

	var g errgroup.Group
	for i := 0; i < m.concurrentMigrators; i++ {
		g.Go(func() error {
			for blob := range blobsToMigrate {
				err := m.blobMigrator.MigrateSingleBlob(blob)
				if err != nil {
					return err
				}
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		m.blobMigrator.Finish(fmt.Sprintf("Error migrating blobs from %s to %s\n", src.Name(), dst.Name()))
		return err
	}

	m.blobMigrator.Finish(fmt.Sprintf("Done migrating blobs from %s to %s\n", src.Name(), dst.Name()))
	return nil
}
