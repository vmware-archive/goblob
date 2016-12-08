package blobstore

import (
	"fmt"
	"io"
	"os"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	nfsclient "github.com/c0-ops/goblob/nfs"
	"github.com/c0-ops/goblob/tar"
)

type Blobstore interface {
	Get(blobPath string, blobID string) (LocalBlob, error)
	Add(blobID string, sourcePath string) (err error)
	GetAll(blobPath string) ([]LocalBlob, error)
}

type blobstore struct {
	bsClient  nfsclient.Client
	fs        boshsys.FileSystem
	extractor tar.CmdExtractor
	logger    boshlog.Logger
	logTag    string
}

func NewBlobstore(bsClient nfsclient.Client, fs boshsys.FileSystem, extractor tar.CmdExtractor, logger boshlog.Logger) Blobstore {
	return &blobstore{
		bsClient:  bsClient,
		fs:        fs,
		extractor: extractor,
		logger:    logger,
		logTag:    "blobstore",
	}
}

func (b *blobstore) Get(blobPath, blobID string) (LocalBlob, error) {

	b.logger.Debug(b.logTag, "Downloading blob %s to %s", blobID, blobPath)

	reader, err := b.bsClient.Get(blobID)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Getting blob %s from blobstore", blobID)
	}

	targetFile, err := b.fs.OpenFile(blobPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Opening file for blob at %s", blobPath)
	}

	_, err = io.Copy(targetFile, reader)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Saving blob to %s", blobPath)
	}

	return NewLocalBlob(blobPath, b.fs, b.logger), nil
}

func (b *blobstore) GetAll(blobPath string) ([]LocalBlob, error) {

	b.logger.Debug(b.logTag, "Downloading blobs")

	extractPath, err := b.bsClient.GetAll(blobPath)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Getting all blobs from blobstore")
	}

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

	b.logger.Debug(b.logTag, "Uploading blob %s to %s", blobID, sourcePath)

	file, err := b.fs.OpenFile(sourcePath, os.O_RDONLY, 0)
	if err != nil {
		return bosherr.WrapErrorf(err, "Opening file for reading %s", sourcePath)
	}
	defer func() {
		if err := file.Close(); err != nil {
			b.logger.Error(b.logTag, "Couldn't close source file", err)
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
