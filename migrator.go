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
		bar := pb.StartNew(len(blobsToMigrate))
		bar.Format("<.- >")
		m.blobMigrator.Init(dst, src, bar)
		return migrate(blobsToMigrate, m.blobMigrator, m.concurrentMigrators)
	}

	return nil
}

func migrate(blobs []*Blob, blobMigrator BlobMigrator, concurrentMigrators int) error {
	var g errgroup.Group

	blobsToMigrate := make(chan *Blob, len(blobs))
	for _, blob := range blobs {
		blobsToMigrate <- blob
	}
	close(blobsToMigrate)

	fmt.Println("Migrating blobs from NFS to S3")

	for i := 0; i < concurrentMigrators; i++ {
		g.Go(func() error {
			for blob := range blobsToMigrate {

				if err := blobMigrator.MigrateSingleBlob(blob); err != nil {
					return err
				}
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		blobMigrator.Finish("Error Migrating blobs from NFS to S3")
		return err
	}
	blobMigrator.Finish("Done Migrating blobs from NFS to S3")
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
