package blobstore

import (
	"os"

	log "code.cloudfoundry.org/lager"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type localblobstore struct {
	fs        boshsys.FileSystem
	logger    log.Logger
}

func NewLocalBlobstore(fs boshsys.FileSystem, logger log.Logger) Blobstore {
	return &localblobstore{
		fs:        fs,
		logger:    logger,
	}
}

func (b *localblobstore) Get(destinationPath, blobID string) (LocalBlob, error) {

	return NewLocalBlob(destinationPath, b.fs, b.logger), nil
}

func (b *localblobstore) GetAll(destinationPath string) ([]LocalBlob, error) {

	b.logger.Info("Downloading blobs")

	b.logger.Debug("Walking files in %s\n" + destinationPath)

	blobs := []LocalBlob{}
	b.fs.Walk(destinationPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			blobs = append(blobs, NewLocalBlob(path, b.fs, b.logger))
		}
		return nil
	})

	return blobs, nil
}

func (b *localblobstore) Add(sourcePath string, blobID string) error {
	return nil
}
