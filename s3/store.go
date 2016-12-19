package s3

import (
	"errors"
	"io"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/c0-ops/goblob"

	awss3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type Store struct {
	session *session.Session
}

func New(config *aws.Config) goblob.Store {
	return &Store{
		session: session.New(config),
	}
}
func (s *Store) List() ([]goblob.Blob, error) {
	return nil, errors.New("not implemented")
}

func (s *Store) bucketName(blob goblob.Blob) string {
	return blob.Path[:strings.Index(blob.Path, "/")]
}

func (s *Store) path(bucketName string, blob goblob.Blob) string {
	return path.Join(blob.Path[len(bucketName)+1:], blob.Filename)
}

func (s *Store) Read(src goblob.Blob) (io.Reader, error) {
	bucketName := s.bucketName(src)
	path := s.path(bucketName, src)
	getObjectOutput, err := awss3.New(s.session).GetObject(&awss3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(path),
	})
	if err != nil {
		return nil, err
	}
	return getObjectOutput.Body, nil

}

func (s *Store) Write(dst goblob.Blob, src io.Reader) error {
	bucketName := s.bucketName(dst)
	path := s.path(bucketName, dst)
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

func (s *Store) createBucket(bucketName string) error {
	var listBucketOutput *awss3.ListBucketsOutput
	var err error
	s3Service := awss3.New(s.session)
	if listBucketOutput, err = s3Service.ListBuckets(&awss3.ListBucketsInput{}); err != nil {
		return err
	}
	for _, bucket := range listBucketOutput.Buckets {
		if *bucket.Name == bucketName {
			return nil
		}
	}
	_, err = s3Service.CreateBucket(&awss3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	return err
}
