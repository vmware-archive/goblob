package s3

import (
	"io"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/minio/minio-go"
)

type Client interface {
	CreateBucket(string, string) error
	//CreateFolder(string) error
	UploadObject(string, string, io.Reader, string) (int64, error)
}

type s3Client struct {
	client *minio.Client
	logger boshlog.Logger
	logTag    string
}

func NewClient(endpoint, accessKey, secretKey string, secure bool, logger boshlog.Logger) (Client, error) {
	mc, err := minio.New(endpoint, accessKey, secretKey, secure)
	if err != nil {
		return nil, err
	}
	return &s3Client{
		client: mc,
		logger: logger,
		logTag: "s3Client",
	}, nil
}

func (c *s3Client) CreateBucket(bucketName string, region string) error {

	c.logger.Info(c.logTag, "Start creating bucket")

	err := c.client.MakeBucket(bucketName, region)
	if err != nil {
		exists, existsErr := c.client.BucketExists(bucketName)
		if existsErr == nil && exists {
			c.logger.Info(c.logTag, "Bucket already exists")
		} else {
			c.logger.Error(c.logTag, "Failed to create bucket %s", err.Error())
			return err
		}
	}
	c.logger.Info(c.logTag, "Done creating bucket")

	return nil
}

func (c *s3Client) UploadObject(bucketName string, objectName string, object io.Reader, contentType string) (int64, error) {

	c.logger.Info(c.logTag, "Start uploading object")

	n, err := c.client.PutObject(bucketName, objectName, object, contentType)
	if err != nil {
		c.logger.Error(c.logTag, "Failed to upload object %s", err.Error())
		return 0, err
	}

	c.logger.Info(c.logTag, "Done uploading object; size: %d", n)

	return n, nil
}
