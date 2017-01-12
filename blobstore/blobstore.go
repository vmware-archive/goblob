package blobstore

import "io"

// Blob is a file in a blob store
type Blob struct {
	Checksum string
	Path     string
}

//go:generate counterfeiter . Blobstore

type Blobstore interface {
	Name() string
	List() ([]*Blob, error)
	Read(src *Blob) (io.ReadCloser, error)
	Checksum(src *Blob) (string, error)
	Write(dst *Blob, src io.Reader) error
	Exists(*Blob) bool
}
