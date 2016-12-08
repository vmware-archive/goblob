package blobstore

import (
	"os"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type localblobstore struct {
	fs     boshsys.FileSystem
	logger boshlog.Logger
	logTag string
}

func NewLocalBlobstore(fs boshsys.FileSystem, logger boshlog.Logger) Blobstore {
	return &localblobstore{
		fs:     fs,
		logger: logger,
		logTag: "localBlobstore",
	}
}

func (b *localblobstore) Get(destinationPath, blobID string) (LocalBlob, error) {
	return NewLocalBlob(destinationPath, b.fs, b.logger), nil
}

func (b *localblobstore) GetAll(blobPath string) ([]LocalBlob, error) {

	b.logger.Debug(b.logTag, "Getting all blobs from %s", blobPath)

	blobs := []LocalBlob{}
	b.fs.Walk(blobPath, func(path string, info os.FileInfo, err error) error {
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
