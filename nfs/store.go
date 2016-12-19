package nfs

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/c0-ops/goblob"
	"github.com/c0-ops/goblob/validation"
	"github.com/cheggaaa/pb"
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
func (s *Store) List() ([]*goblob.Blob, error) {
	var blobs []*goblob.Blob
	walk := func(path string, info os.FileInfo, e error) error {
		if !info.IsDir() && info.Name() != ".nfs_test" {
			filePath := path[len(s.path)-1 : len(path)-len(info.Name())-1]
			blobs = append(blobs, &goblob.Blob{
				Filename: info.Name(),
				Path:     filePath,
			})
		}
		return e
	}
	if err := filepath.Walk(s.path, walk); err != nil {
		return nil, err
	}

	fmt.Println("Getting list of files from NFS")
	bar := pb.StartNew(len(blobs))
	bar.Format("<.- >")
	for _, blob := range blobs {
		checksum, err := validation.Checksum(path.Join(s.path, blob.Path, blob.Filename))
		if (err) != nil {
			return nil, err
		}
		blob.Checksum = checksum
		bar.Increment()
	}
	bar.FinishPrint("Done Getting list of files from NFS")
	return blobs, nil
}

func (s *Store) Read(src *goblob.Blob) (io.Reader, error) {
	return os.Open(path.Join(s.path, src.Path, src.Filename))
}
func (s *Store) Write(dst *goblob.Blob, src io.Reader) error {
	return errors.New("writing to the NFS store is not supported")
}
