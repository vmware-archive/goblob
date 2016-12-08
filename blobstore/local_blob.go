package blobstore

import (
	"fmt"

	"code.cloudfoundry.org/lager"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

// LocalBlob represents a local copy of a blob retrieved from the blobstore
type LocalBlob interface {
	// Path returns the path to the local copy of the blob
	Path() string
	// Delete removes the local copy of the blob (does not effect the blobstore)
	Delete() error
	// DeleteSilently removes the local copy of the blob (does not effect the blobstore), logging instead of returning an error.
	DeleteSilently()
}

type localBlob struct {
	path   string
	fs     boshsys.FileSystem
	logger lager.Logger
}

func NewLocalBlob(path string, fs boshsys.FileSystem, logger lager.Logger) LocalBlob {
	return &localBlob{
		path:   path,
		fs:     fs,
		logger: logger,
	}
}

func (b *localBlob) Path() string {
	return b.path
}

func (b *localBlob) Delete() error {
	err := b.fs.RemoveAll(b.path)
	if err != nil {
		return bosherr.WrapErrorf(err, "Deleting local blob '%s'", b.path)
	}
	return nil
}

func (b *localBlob) DeleteSilently() {
	err := b.Delete()
	if err != nil {
		data := lager.Data{"event": "Failed to delete local blob"}
		b.logger.Error("delete-silently",err, data)
	}
}

func (b *localBlob) String() string {
	return fmt.Sprintf("localBlob{path: '%s'}", b.path)
}
