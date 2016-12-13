package mock

import (
	"io"

	"github.com/c0-ops/goblob"
)

type CloudFoundry struct {
	ListFn  func() ([]goblob.Blob, error)
	ReadFn  func(src goblob.Blob) (io.Reader, error)
	WriteFn func(dst goblob.Blob, src io.Writer) error
}

func (c *CloudFoundry) Identifier() string {
	return "12345678"
}

func (c *CloudFoundry) EnableBits() error {
	return nil
}

func (c *CloudFoundry) DisableBits() error {
	return nil
}

func (c *CloudFoundry) Reconfigure(dst goblob.Store) error {
	return nil
}

func (c *CloudFoundry) List() ([]goblob.Blob, error) {
	if c.ListFn != nil {
		return c.ListFn()
	}
	return nil, nil
}

func (c *CloudFoundry) Read(src goblob.Blob) (io.Reader, error) {
	if c.ReadFn != nil {
		return c.ReadFn(src)
	}
	return nil, nil
}

func (c *CloudFoundry) Write(dst goblob.Blob, src io.Writer) error {
	if c.WriteFn != nil {
		return c.WriteFn(dst, src)
	}
	return nil
}
