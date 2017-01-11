package goblob

import (
	"fmt"
	"path"
)

// Blob is a file in a blob store
type Blob struct {
	Filename string
	Checksum string
	Path     string
}

//go:generate counterfeiter . BlobMigrator

type BlobMigrator interface {
	Migrate(blob *Blob) error
}

type blobMigrator struct {
	dst Blobstore
	src Blobstore
}

func NewBlobMigrator(dst Blobstore, src Blobstore) BlobMigrator {
	return &blobMigrator{
		dst: dst,
		src: src,
	}
}

func (m *blobMigrator) Migrate(blob *Blob) error {
	reader, err := m.src.Read(blob)
	if err != nil {
		return fmt.Errorf("error at %s: %s", path.Join(blob.Path, blob.Filename), err)
	}
	defer reader.Close()

	err = m.dst.Write(blob, reader)
	if err != nil {
		return fmt.Errorf("error at %s: %s", path.Join(blob.Path, blob.Filename), err)
	}

	checksum, err := m.dst.Checksum(blob)
	if err != nil {
		return fmt.Errorf("error at %s: %s", path.Join(blob.Path, blob.Filename), err)
	}

	if checksum != blob.Checksum {
		return fmt.Errorf(
			"error at %s: checksum [%s] does not match [%s]",
			path.Join(blob.Path, blob.Filename),
			checksum,
			blob.Checksum,
		)
	}

	return nil
}
