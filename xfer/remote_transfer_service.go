package xfer

import (
	"path"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	"github.com/c0-ops/goblob/blobstore"
	"github.com/c0-ops/goblob/s3"
)

type rXferSvc struct {
	xferSvc
	client    s3.Client
	blobstore blobstore.Blobstore
	logger    boshlog.Logger
	logTag    string
}

func NewRemoteTransferService(svc xferSvc, s3client s3.Client, bs blobstore.Blobstore, logger boshlog.Logger) rXferSvc {
	return rXferSvc{
		xferSvc:   svc,
		client:    s3client,
		blobstore: bs,
		logger:    logger,
		logTag:    "rXferSvc",
	}
}

func (s rXferSvc) Transfer(buckets []string, destinationPath string) error {
	for _, bucketName := range buckets {
		if strings.Trim(bucketName, "ignore") != "" {
			blobPath := path.Join(destinationPath, bucketName)
			blobs, err := s.blobstore.GetAll(blobPath)
			if err != nil {
				s.logger.Error(s.logTag, "Failed to get all blobs %v", err)
				return bosherr.WrapError(err, "Transfering blobs to s3")
			}
			s.xferSvc.populateS3blobstore(bucketName, blobs)
		}
	}
	return nil
}
