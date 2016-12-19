package nfs

import (
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/c0-ops/goblob"
	"github.com/c0-ops/goblob/validation"
)

// Store is an NFS blob store
type Store struct {
	path string
}

// New creates an NFS blob store
func New(path string) goblob.Store {
	return &Store{
		path: path,
	}
}

// List fetches a list of files with checksums
func (s *Store) List() ([]goblob.Blob, error) {
	var blobs []goblob.Blob
	walk := func(path string, info os.FileInfo, e error) error {
		if !info.IsDir() {
			filePath := path[len(s.path)-1 : len(path)-len(info.Name())-1]
			checksum, err := validation.Checksum(path)
			if (err) != nil {
				return err
			}
			blobs = append(blobs, goblob.Blob{
				Filename: info.Name(),
				Path:     filePath,
				Checksum: checksum,
			})
		}

		return e
	}
	if err := filepath.Walk(s.path, walk); err != nil {
		return nil, err
	}
	return blobs, nil
}

func (s *Store) Read(src goblob.Blob) (io.Reader, error) {
	return os.Open(path.Join(s.path, src.Path, src.Filename))
}
func (s *Store) Write(dst goblob.Blob, src io.Reader) error {
	return errors.New("writing to the NFS store is not supported")
}
