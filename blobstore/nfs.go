package blobstore

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/c0-ops/goblob/validation"
	"github.com/cheggaaa/pb"
	"golang.org/x/sync/errgroup"
)

type nfsStore struct {
	path string
}

// NewNFS creates an NFS blobstore
func NewNFS(path string) Blobstore {
	return &nfsStore{
		path: path,
	}
}

func (s *nfsStore) Name() string {
	return "NFS"
}

// List fetches a list of files with checksums
func (s *nfsStore) List() ([]*Blob, error) {
	var blobs []*Blob
	walk := func(path string, info os.FileInfo, e error) error {
		if !info.IsDir() && info.Name() != ".nfs_test" {
			relPath := path[len(s.path)+1:]
			blobs = append(blobs, &Blob{
				Path: relPath,
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

func (s *nfsStore) processBlobsForChecksums(blobs []*Blob) error {

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

func (s *nfsStore) Checksum(src *Blob) (string, error) {
	return validation.Checksum(path.Join(s.path, src.Path))
}

func (s *nfsStore) Read(src *Blob) (io.ReadCloser, error) {
	return os.Open(path.Join(s.path, src.Path))
}
func (s *nfsStore) Write(dst *Blob, src io.Reader) error {
	return errors.New("writing to the NFS store is not supported")
}

func (s *nfsStore) Exists(blob *Blob) bool {
	checksum, err := s.Checksum(blob)
	if err != nil {
		return false
	}

	return checksum == blob.Checksum
}

func (s *nfsStore) NewBucketIterator(folder string) (BucketIterator, error) {
	blobCh := make(chan *Blob)
	doneCh := make(chan struct{})
	errCh := make(chan error)

	actualPath := filepath.Join(s.path, folder)

	files, err := ioutil.ReadDir(actualPath)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return &nfsBucketIterator{}, nil
	}

	iterator := &nfsBucketIterator{
		blobCh:    blobCh,
		doneCh:    doneCh,
		errCh:     errCh,
		blobCount: uint(len(files)),
	}

	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || info.Name() == ".nfs_test" {
			return nil
		}

		select {
		case <-doneCh:
			doneCh = nil
			return ErrIteratorAborted
		default:
			blob := &Blob{
				Path: strings.TrimPrefix(path, s.path+string(os.PathSeparator)),
			}

			blobCh <- blob

			return nil
		}
	}

	go func() {
		errCh <- filepath.Walk(actualPath, walkFn)
		close(blobCh)
	}()

	return iterator, nil
}
