package s3

import (
	"io"

	"code.cloudfoundry.org/lager"
	"github.com/minio/minio-go"
)

type Client interface {
	CreateBucket(string, string) error
	//CreateFolder(string) error
	UploadObject(string, string, io.Reader, string) (int64, error)
}

type s3Client struct {
	client *minio.Client
	logger lager.Logger
}

func NewClient(endpoint, accessKey, secretKey string, secure bool, logger lager.Logger) (Client, error) {
	mc, err := minio.New(endpoint, accessKey, secretKey, secure)
	if err != nil {
		return nil, err
	}
	return &s3Client{
		client: mc,
		logger: logger,
	}, nil
}

//func (c *s3Client) CreateFolder(folder string) error {
//
//}

func (c *s3Client) CreateBucket(bucketName string, region string) error {
	logData := lager.Data{
		"bucket_name": bucketName,
		"region":      region,
	}

	c.logInfo("s3client.create-bucket", "starting", logData)

	err := c.client.MakeBucket(bucketName, region)
	if err != nil {
		exists, existsErr := c.client.BucketExists(bucketName)
		if existsErr == nil && exists {
			c.logInfo("s3client.create-bucket", "already-exists", logData)
		} else {
			c.logError("s3client.create-bucket", err, logData)
			return err
		}
	}
	c.logInfo("s3client.create-bucket", "done", logData)

	return nil
}

func (c *s3Client) UploadObject(bucketName string, objectName string, object io.Reader, contentType string) (int64, error) {
	logData := lager.Data{
		"bucket_name":  bucketName,
		"object_name":  objectName,
		"content_type": contentType,
		"size":         0,
	}

	c.logInfo("s3client.upload-object", "uploading", logData)

	n, err := c.client.PutObject(bucketName, objectName, object, contentType)
	if err != nil {
		c.logError("s3client.upload-object", err, logData)
		return 0, err
	}
	logData["size"] = n

	c.logInfo("s3client.upload-object", "done", logData)

	return n, nil
}

func (c *s3Client) logInfo(action, event string, data lager.Data) {
	data["event"] = event
	c.logger.Info(action, data)
}

func (c *s3Client) logError(action string, err error, data lager.Data) {
	data["event"] = "failed"
	c.logger.Error(action, err, data)
}
