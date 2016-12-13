package mock

import (
	"io"

	"github.com/c0-ops/goblob"
)

type S3 struct {
	ListFn  func() ([]goblob.Blob, error)
	ReadFn  func(src goblob.Blob) (io.Reader, error)
	WriteFn func(dst goblob.Blob, src io.Writer) error
}

func (c *S3) List() ([]goblob.Blob, error) {
	if c.ListFn != nil {
		return c.ListFn()
	}
	return nil, nil
}

func (c *S3) Read(src goblob.Blob) (io.Reader, error) {
	if c.ReadFn != nil {
		return c.ReadFn(src)
	}
	return nil, nil
}

func (c *S3) Write(dst goblob.Blob, src io.Writer) error {
	if c.WriteFn != nil {
		return c.WriteFn(dst, src)
	}
	return nil
}
