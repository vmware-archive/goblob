package goblob

import "fmt"

// Blob is a file in a blob store
type Blob struct {
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
		return fmt.Errorf("error at %s: %s", blob.Path, err)
	}
	defer reader.Close()

	err = m.dst.Write(blob, reader)
	if err != nil {
		return fmt.Errorf("error at %s: %s", blob.Path, err)
	}

	checksum, err := m.dst.Checksum(blob)
	if err != nil {
		return fmt.Errorf("error at %s: %s", blob.Path, err)
	}

	if checksum != blob.Checksum {
		return fmt.Errorf(
			"error at %s: checksum [%s] does not match [%s]",
			blob.Path,
			checksum,
			blob.Checksum,
		)
	}

	return nil
}
