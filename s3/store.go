package s3

import (
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	awss3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/c0-ops/goblob"
	"github.com/c0-ops/goblob/validation"
	"github.com/cheggaaa/pb"
	"github.com/xchapter7x/lo"
)

var (
	buckets = []string{"cc-buildpacks", "cc-droplets", "cc-packages", "cc-resources"}
)

type Store struct {
	session    *session.Session
	identifier string
}

func New(identifier, awsAccessKey, awsSecretKey, region, endpoint string) goblob.Store {
	return &Store{
		session: session.New(&aws.Config{
			Region:           aws.String(region),
			Credentials:      credentials.NewStaticCredentials(awsAccessKey, awsSecretKey, ""),
			Endpoint:         aws.String(endpoint),
			DisableSSL:       aws.Bool(true),
			S3ForcePathStyle: aws.Bool(true),
		}),
		identifier: identifier,
	}
}
func (s *Store) List() ([]*goblob.Blob, error) {
	var blobs []*goblob.Blob
	s3Service := awss3.New(s.session)
	for _, bucket := range buckets {
		bucketName := bucket + "-" + s.identifier
		bucketExists, err := s.doesBucketExist(bucketName)
		if err != nil {
			return nil, err
		}
		if bucketExists {
			listObjectsOutput, err := s3Service.ListObjects(&awss3.ListObjectsInput{
				Bucket: aws.String(bucketName),
			})
			if err != nil {
				return nil, err
			}
			fmt.Println("Getting list of files from S3", bucketName)
			bar := pb.StartNew(len(listObjectsOutput.Contents))
			bar.Format("<.- >")
			for _, item := range listObjectsOutput.Contents {
				thePath := *item.Key
				fileName := thePath[strings.LastIndex(thePath, "/")+1:]
				filePath := thePath[:strings.LastIndex(thePath, "/")]
				blobPath := path.Join(bucket, filePath)
				blob := &goblob.Blob{
					Filename: fileName,
					Path:     blobPath,
				}
				reader, err := s.Read(blob)
				if err != nil {
					return nil, err
				}
				checksum, err := validation.ChecksumReader(reader)
				if err != nil {
					return nil, err
				}
				blob.Checksum = checksum
				blobs = append(blobs, blob)
				bar.Increment()
			}
			bar.FinishPrint(fmt.Sprintf("Done Getting list of files from S3 %s", bucketName))
		}
	}
	return blobs, nil
}

func (s *Store) bucketName(blob *goblob.Blob) string {
	return blob.Path[:strings.Index(blob.Path, "/")] + "-" + s.identifier
}

func (s *Store) path(blob *goblob.Blob) string {
	bucketName := blob.Path[:strings.Index(blob.Path, "/")]
	return path.Join(blob.Path[len(bucketName)+1:], blob.Filename)
}

func (s *Store) Read(src *goblob.Blob) (io.Reader, error) {
	bucketName := s.bucketName(src)
	path := s.path(src)
	lo.G.Debug("Getting", path, "from bucket", bucketName)
	getObjectOutput, err := awss3.New(s.session).GetObject(&awss3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(path),
	})
	if err != nil {
		return nil, err
	}
	return getObjectOutput.Body, nil

}

func (s *Store) Write(dst *goblob.Blob, src io.Reader) error {
	bucketName := s.bucketName(dst)
	path := s.path(dst)
	if err := s.createBucket(bucketName); err != nil {
		return err
	}
	uploader := s3manager.NewUploader(s.session)
	_, err := uploader.Upload(&s3manager.UploadInput{
		Body:   src,
		Bucket: aws.String(bucketName),
		Key:    aws.String(path),
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) doesBucketExist(bucketName string) (bool, error) {
	var listBucketOutput *awss3.ListBucketsOutput
	var err error
	s3Service := awss3.New(s.session)
	if listBucketOutput, err = s3Service.ListBuckets(&awss3.ListBucketsInput{}); err != nil {
		return false, err
	}
	for _, bucket := range listBucketOutput.Buckets {
		if *bucket.Name == bucketName {
			return true, nil
		}
	}
	return false, nil
}

func (s *Store) createBucket(bucketName string) error {

	s3Service := awss3.New(s.session)
	bucketExists, err := s.doesBucketExist(bucketName)
	if err != nil {
		return err
	}
	if bucketExists {
		return nil
	}
	_, err = s3Service.CreateBucket(&awss3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})

	return err

}
