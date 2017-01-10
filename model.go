package goblob

import "io"

// CloudFoundry is a Cloud Foundry deployment
//go:generate counterfeiter -o ./mock/fakecloudfoundry.go . CloudFoundry
type CloudFoundry interface {
	Identifier() string
	EnableBits() error
	DisableBits() error
	Reconfigure(dst Store) error
	Store() (Store, error)
}

// Store lists, reads, and writes blobs
//go:generate counterfeiter -o ./mock/fakestore.go . Store
type Store interface {
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

// Copier moves files from the src to the destination
type Copier interface {
	Copy(dst io.Writer, src io.Reader) (int64, error)
}

// Migrator moves blobs from one Store to another
type Migrator interface {
	Migrate(dst Store, src Store) error
}
