package blobstore

import (
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	"code.cloudfoundry.org/lager"
)

type LocalBlobstoreFactory interface {
	NewBlobstore(logger lager.Logger) (Blobstore, error)
}

type localBlobstoreFactory struct {
	fs     boshsys.FileSystem
	logger lager.Logger
}

func NewLocalBlobstoreFactory(fs boshsys.FileSystem, logger lager.Logger) LocalBlobstoreFactory {
	return localBlobstoreFactory{
		fs:     fs,
		logger: logger,
	}
}

func (f localBlobstoreFactory) NewBlobstore(logger lager.Logger) (Blobstore, error) {
	return NewLocalBlobstore(f.fs, f.logger), nil
}
