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
	"golang.org/x/sync/errgroup"
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
			filePath := path[len(s.path)+1 : len(path)-len(info.Name())-1]
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
	if err := s.processBlobsForChecksums(blobs); err != nil {
		return nil, err
	}
	return blobs, nil
}

func (s *Store) processBlobsForChecksums(blobs []*goblob.Blob) error {

	fmt.Println("Getting list of files from NFS")
	bar := pb.StartNew(len(blobs))
	bar.Format("<.- >")

	var g errgroup.Group
	for _, blob := range blobs {
		blob := blob
		g.Go(func() error {
			checksum, err := s.Checksum(blob)
			if (err) != nil {
				return err
			}
			blob.Checksum = checksum
			bar.Increment()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return err
	}

	bar.FinishPrint("Done Getting list of files from NFS")
	return nil
}

func (s *Store) Checksum(src *goblob.Blob) (string, error) {
	return validation.Checksum(path.Join(s.path, src.Path, src.Filename))
}

func (s *Store) Read(src *goblob.Blob) (io.ReadCloser, error) {
	return os.Open(path.Join(s.path, src.Path, src.Filename))
}
func (s *Store) Write(dst *goblob.Blob, src io.Reader) error {
	return errors.New("writing to the NFS store is not supported")
}

func (s *Store) Exists(blob *goblob.Blob) bool {
	checksum, err := s.Checksum(blob)
	if err != nil {
		return false
	}

	return checksum == blob.Checksum
}
