package goblob

import "io"

// Store lists, reads, and writes blobs
//go:generate counterfeiter -o ./mock/fakestore.go . Store
type Store interface {
	Name() string
	List() ([]*Blob, error)
	Read(src *Blob) (io.ReadCloser, error)
	Checksum(src *Blob) (string, error)
	Write(dst *Blob, src io.Reader) error
	Exists(*Blob) bool
}

// Blob is a file in a blob store
type Blob struct {
	Filename string
	Checksum string
	Path     string
}
