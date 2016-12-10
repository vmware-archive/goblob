package xfer

import (
	"os"
	"path"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	"github.com/c0-ops/goblob/blobstore"
	"github.com/c0-ops/goblob/s3"
)

type xferSvc struct {
	client    s3.Client
	blobstore blobstore.Blobstore
	logger    boshlog.Logger
	logTag    string
}

func NewTransferService(s3client s3.Client, bs blobstore.Blobstore, logger boshlog.Logger) xferSvc {
	return xferSvc{
		client:    s3client,
		blobstore: bs,
		logger:    logger,
		logTag:    "xferSvc",
	}
}

func (s xferSvc) Transfer(buckets []string, dest string) error {
	for _, bucketName := range buckets {
		if strings.Trim(bucketName, "ignore") != "" {
			blobPath := path.Join(dest, bucketName)
			blobs, err := s.blobstore.GetAll(blobPath)
			if err != nil {
				s.logger.Error(s.logTag, "Failed to get all blobs %v", err)
				return bosherr.WrapError(err, "Transfering blobs to s3")
			}
			s.populateS3blobstore(bucketName, blobs)
		}
	}
	return nil
}

func (s xferSvc) populateS3blobstore(bucketName string, blobs []blobstore.LocalBlob) {
	err := s.createBucket(bucketName)
	if err != nil {
		s.logger.Error(s.logTag, "Failed to create bucket %v", err)
	}
	for _, blob := range blobs {
		s.logger.Debug(s.logTag, "Uploading blob %s", blob.Path())
		err = s.uploadObject(bucketName, blob)
		if err != nil {
			s.logger.Error(s.logTag, "Failed to upload blobs %s %v", blob.Path())
		}
	}
}

func (s xferSvc) createBucket(bucketName string) error {
	return s.client.CreateBucket(bucketName)
}

func (s xferSvc) uploadObject(bucketName string, blob blobstore.LocalBlob) error {
	object, err := os.Open(blob.Path())
	if err != nil {
		return err
	}

	_, file := split(blob.Path())

	_, err = s.client.UploadObject(bucketName, file, object, "")
	return err
}

func split(path string) (dir, file string) {
	i := strings.LastIndex(path, "/")
	return path[:i+1], path[i+1:]
}
