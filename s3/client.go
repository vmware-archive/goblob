package s3

import (
	"fmt"
	"io"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/minio/minio-go"
)

type Client interface {
	CreateBucket(bucketName string) error
	UploadObject(bucketName string, objectName string, object io.Reader, contentType string) (int64, error)
}

type Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	UseSSL          bool
}

type s3Client struct {
	client *minio.Client
	config Config
	logger boshlog.Logger
	logTag string
}

func NewClient(config Config, logger boshlog.Logger) (Client, error) {
	mc, err := minio.New(config.Endpoint, config.AccessKeyID, config.SecretAccessKey, config.UseSSL)
	if err != nil {
		return nil, err
	}
	return &s3Client{
		client: mc,
		config: config,
		logger: logger,
		logTag: "s3Client",
	}, nil
}

func (c *s3Client) CreateBucket(bucketName string) error {

	c.logger.Info(c.logTag, "Start creating bucket")

	fmt.Printf("Making bucket %s in region %s", bucketName, c.config.Region)

	err := c.client.MakeBucket(bucketName, c.config.Region)
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
