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
	endpoint, accessKeyID, secretAccessKey, region string
	blobstore                                      blobstore.Blobstore
	logger                                         boshlog.Logger
	logTag                                         string
}

func NewTransferService(endpoint, accessKeyID, secretAccessKey, region string, bs blobstore.Blobstore, logger boshlog.Logger) xferSvc {
	return xferSvc{
		endpoint:        endpoint,
		accessKeyID:     accessKeyID,
		secretAccessKey: secretAccessKey,
		region:          region,
		blobstore:       bs,
		logger:          logger,
		logTag:          "xferSvc",
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
			s.populateS3blobstore(s.endpoint, s.accessKeyID, s.secretAccessKey, bucketName, s.region, blobs)
		}
	}
	return nil
}

func (s xferSvc) populateS3blobstore(endpoint, accessKeyID, secretAccessKey, bucketName, region string, blobs []blobstore.LocalBlob) {
	err := s.createBucket(endpoint, accessKeyID, secretAccessKey, bucketName, region)
	if err != nil {
		s.logger.Error(s.logTag, "Failed to create bucket %v", err)
	}
	for _, blob := range blobs {
		s.logger.Debug(s.logTag, "Uploading blob %s", blob.Path())
		err = s.uploadObject(endpoint, accessKeyID, secretAccessKey, bucketName, blob)
		if err != nil {
			s.logger.Error(s.logTag, "Failed to upload blobs %s %v", blob.Path())
		}
	}
}

func (s xferSvc) createBucket(endpoint, accessKeyID, secretAccessKey, bucketName, region string) error {
	client, err := s3.NewClient(endpoint, accessKeyID, secretAccessKey, false, s.logger)
	if err != nil {
		return err
	}
	err = client.CreateBucket(bucketName, region)
	return err
}

func (s xferSvc) uploadObject(endpoint, accessKeyID, secretAccessKey, bucketName string, blob blobstore.LocalBlob) error {
	client, err := s3.NewClient(endpoint, accessKeyID, secretAccessKey, false, s.logger)
	if err != nil {
		return err
	}
	//object, _ := fs.OpenFile(blob.Path(), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	object, err := os.Open(blob.Path())
	if err != nil {
		return err
	}

	_, file := split(blob.Path())

	_, err = client.UploadObject(bucketName, file, object, "")
	return err
}

func split(path string) (dir, file string) {
	i := strings.LastIndex(path, "/")
	return path[:i+1], path[i+1:]
}
