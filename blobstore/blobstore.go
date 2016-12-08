package blobstore

import (
	"io"
	"os"

	log "code.cloudfoundry.org/lager"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	nfsclient "github.com/pivotal-customer0/cfblobmigrator/nfs"
	"github.com/pivotal-customer0/cfblobmigrator/tar"
	"fmt"
)

type Blobstore interface {
	Get(destinationPath string, blobID string) (LocalBlob, error)
	Add(blobID string, sourcePath string) (err error)
	GetAll(destinationPath string) ([]LocalBlob, error)
}

type blobstore struct {
	bsClient  nfsclient.Client
	fs        boshsys.FileSystem
	extractor tar.CmdExtractor
	logger    log.Logger
}

func NewBlobstore(bsClient nfsclient.Client, fs boshsys.FileSystem, extractor tar.CmdExtractor, logger log.Logger) Blobstore {
	return &blobstore{
		bsClient:  bsClient,
		fs:        fs,
		extractor: extractor,
		logger:    logger,
	}
}

func (b *blobstore) Get(destinationPath, blobID string) (LocalBlob, error) {

	logData := log.Data{
		"blob_id":          blobID,
		"destination_path": destinationPath,
	}

	b.logger.Debug("Downloading blob %s to %s", logData)

	reader, err := b.bsClient.Get(blobID)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Getting blob %s from blobstore", blobID)
	}

	targetFile, err := b.fs.OpenFile(destinationPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Opening file for blob at %s", destinationPath)
	}

	_, err = io.Copy(targetFile, reader)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Saving blob to %s", destinationPath)
	}

	return NewLocalBlob(destinationPath, b.fs, b.logger), nil
}

func (b *blobstore) GetAll(destinationPath string) ([]LocalBlob, error) {

	b.logger.Debug("Downloading blobs")

	extractPath, err := b.bsClient.GetAll(destinationPath)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Getting all blobs from blobstore")
	}
	//
	//fmt.Printf("Opening file for blobs at %s\n", destinationPath)
	//
	//targetFile, err := b.fs.OpenFile(destinationPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	//if err != nil {
	//	return nil, bosherr.WrapErrorf(err, "Opening file for blobs at %s", destinationPath)
	//}
	//
	//fmt.Printf("Saving blobs to %s\n", destinationPath)
	//
	//_, err = io.Copy(targetFile, reader)
	//if err != nil {
	//	return nil, bosherr.WrapErrorf(err, "Saving blobs to %s", destinationPath)
	//}

	//extractPath, err := b.extractor.Extract(destinationPath)
	fmt.Printf("Walking files in %s\n", extractPath)

	blobs := []LocalBlob{}
	b.fs.Walk(extractPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			blobs = append(blobs, NewLocalBlob(path, b.fs, b.logger))
		}
		return nil
	})

	return blobs, nil
}

func (b *blobstore) Add(sourcePath string, blobID string) error {

	logData := log.Data{
		"blob_id":     blobID,
		"source_path": sourcePath,
	}

	b.logger.Debug("Uploading blob %s to %s", logData)

	file, err := b.fs.OpenFile(sourcePath, os.O_RDONLY, 0)
	if err != nil {
		return bosherr.WrapErrorf(err, "Opening file for reading %s", sourcePath)
	}
	defer func() {
		if err := file.Close(); err != nil {
			b.logger.Error("Couldn't close source file", err, logData)
		}
	}()

	fileInfo, err := file.Stat()
	if err != nil {
		return bosherr.WrapErrorf(err, "Getting fileInfo from %s", sourcePath)
	}

	err = b.bsClient.Put(blobID, file, fileInfo.Size())
	if err != nil {
		return bosherr.WrapErrorf(err, "Putting file '%s' into blobstore (via Client) as blobID '%s'", sourcePath, blobID)
	}

	return nil
}
