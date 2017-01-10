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
	session             *session.Session
	identifier          string
	useMultipartUploads bool
}

func New(identifier, awsAccessKey, awsSecretKey, region, endpoint string, useMultipartUploads bool) goblob.Store {
	return &Store{
		session: session.New(&aws.Config{
			Region:           aws.String(region),
			Credentials:      credentials.NewStaticCredentials(awsAccessKey, awsSecretKey, ""),
			Endpoint:         aws.String(endpoint),
			DisableSSL:       aws.Bool(true),
			S3ForcePathStyle: aws.Bool(true),
		}),
		identifier:          identifier,
		useMultipartUploads: useMultipartUploads,
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

				checksum, err := s.checksumFromMetadata(blob)
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

func (s *Store) Checksum(src *goblob.Blob) (string, error) {
	if s.useMultipartUploads {
		getObjectOutput, err := awss3.New(s.session).GetObject(&awss3.GetObjectInput{
			Bucket: aws.String(s.bucketName(src)),
			Key:    aws.String(s.path(src)),
		})
		if err != nil {
			return "", err
		}
		defer getObjectOutput.Body.Close()
		return validation.ChecksumReader(getObjectOutput.Body)
	} else {
		return s.checksumFromETAG(src)
	}
}

func (s *Store) checksumFromETAG(src *goblob.Blob) (string, error) {
	headObjectOutput, err := awss3.New(s.session).HeadObject(&awss3.HeadObjectInput{
		Bucket: aws.String(s.bucketName(src)),
		Key:    aws.String(s.path(src)),
	})
	if err != nil {
		return "", err
	}
	return strings.Replace(*headObjectOutput.ETag, "\"", "", -1), nil
}

func (s *Store) checksumFromMetadata(src *goblob.Blob) (string, error) {
	headObjectOutput, err := awss3.New(s.session).HeadObject(&awss3.HeadObjectInput{
		Bucket: aws.String(s.bucketName(src)),
		Key:    aws.String(s.path(src)),
	})
	if err != nil {
		return "", err
	}
	value, ok := headObjectOutput.Metadata["Checksum"]
	if ok {
		return *value, nil
	} else {
		return "", nil
	}
}

func (s *Store) Read(src *goblob.Blob) (io.ReadCloser, error) {
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
	metadataMap := map[string]*string{
		"Checksum": &dst.Checksum,
	}
	if s.useMultipartUploads {
		uploader := s3manager.NewUploader(s.session)
		_, err := uploader.Upload(&s3manager.UploadInput{
			Body:     src,
			Bucket:   aws.String(bucketName),
			Key:      aws.String(path),
			Metadata: metadataMap,
		}, func(u *s3manager.Uploader) {
			u.PartSize = 10 * 1024 * 1024 // 10MB part size
			u.Concurrency = 20
		})
		if err != nil {
			return err
		}
	} else {
		_, err := awss3.New(s.session).PutObject(&awss3.PutObjectInput{
			Body:     aws.ReadSeekCloser(src),
			Bucket:   aws.String(bucketName),
			Key:      aws.String(path),
			Metadata: metadataMap,
		})
		if err != nil {
			return err
		}
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

func (s *Store) Exists(blob *goblob.Blob) bool {
	checksum, err := s.Checksum(blob)
	if err != nil {
		return false
	}

	return checksum == blob.Checksum
}
