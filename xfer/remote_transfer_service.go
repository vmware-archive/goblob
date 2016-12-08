package xfer

import (
	"path"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	"github.com/c0-ops/goblob/blobstore"
)

type rXferSvc struct {
	xferSvc
	endpoint, accessKeyID, secretAccessKey, region string
	blobstore                                      blobstore.Blobstore
	logger                                         boshlog.Logger
	logTag                                         string
}

func NewRemoteTransferService(svc xferSvc, endpoint, accessKeyID, secretAccessKey, region string, bs blobstore.Blobstore, logger boshlog.Logger) rXferSvc {
	return rXferSvc{
		xferSvc:         svc,
		endpoint:        endpoint,
		accessKeyID:     accessKeyID,
		secretAccessKey: secretAccessKey,
		region:          region,
		blobstore:       bs,
		logger:          logger,
		logTag:          "transferService",
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
			s.xferSvc.populateS3blobstore(s.endpoint, s.accessKeyID, s.secretAccessKey, bucketName, s.region, blobs)
		}
	}
	return nil
}
