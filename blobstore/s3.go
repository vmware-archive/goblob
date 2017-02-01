package blobstore

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	awss3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/cheggaaa/pb"
	"github.com/pivotalservices/goblob/validation"
	"github.com/xchapter7x/lo"
)

var (
	buckets = []string{"cc-buildpacks", "cc-droplets", "cc-packages", "cc-resources"}
)

type s3Store struct {
	session             *session.Session
	identifier          string
	useMultipartUploads bool
	bucketMapping       map[string]string
}

func NewS3(
	identifier string,
	awsAccessKey string,
	awsSecretKey string,
	region string,
	endpoint string,
	useMultipartUploads bool,
	disableSSL bool,
	buildpacksBucketName string,
	dropletsBucketName string,
	packagesBucketName string,
	resourcesBucketName string,
) Blobstore {
	bucketMapping := map[string]string{
		"cc-buildpacks": buildpacksBucketName,
		"cc-resources":  resourcesBucketName,
		"cc-droplets":   dropletsBucketName,
		"cc-packages":   packagesBucketName,
	}
	return &s3Store{
		session: session.New(&aws.Config{
			Region:           aws.String(region),
			Credentials:      credentials.NewStaticCredentials(awsAccessKey, awsSecretKey, ""),
			Endpoint:         aws.String(endpoint),
			DisableSSL:       aws.Bool(disableSSL),
			S3ForcePathStyle: aws.Bool(true),
		}),
		identifier:          identifier,
		useMultipartUploads: useMultipartUploads,
		bucketMapping:       bucketMapping,
	}
}

func (s *s3Store) Name() string {
	return "S3"
}

func (s *s3Store) destBucketName(bucket string) string {
	if name, ok := s.bucketMapping[bucket]; ok && name != "" {
		return name
	}
	return bucket + s.identifier
}

func (s *s3Store) List() ([]*Blob, error) {
	var blobs []*Blob
	s3Service := awss3.New(s.session)
	for _, bucket := range buckets {
		bucketName := s.destBucketName(bucket)
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
				blob := &Blob{
					Path: filepath.Join(bucket, *item.Key),
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

func (s *s3Store) bucketName(blob *Blob) string {
	return s.destBucketName(blob.Path[:strings.Index(blob.Path, "/")])
}

func (s *s3Store) path(blob *Blob) string {
	bucketName := blob.Path[:strings.Index(blob.Path, "/")]
	return blob.Path[len(bucketName)+1:]
}

func (s *s3Store) Checksum(src *Blob) (string, error) {
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
	}

	return s.checksumFromETAG(src)
}

func (s *s3Store) checksumFromETAG(src *Blob) (string, error) {
	headObjectOutput, err := awss3.New(s.session).HeadObject(&awss3.HeadObjectInput{
		Bucket: aws.String(s.bucketName(src)),
		Key:    aws.String(s.path(src)),
	})
	if err != nil {
		return "", err
	}
	return strings.Replace(*headObjectOutput.ETag, "\"", "", -1), nil
}

func (s *s3Store) checksumFromMetadata(src *Blob) (string, error) {
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
	}

	return "", nil
}

func (s *s3Store) Read(src *Blob) (io.ReadCloser, error) {
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

func (s *s3Store) Write(dst *Blob, src io.Reader) error {
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

func (s *s3Store) doesBucketExist(bucketName string) (bool, error) {
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

func (s *s3Store) createBucket(bucketName string) error {

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

func (s *s3Store) Exists(blob *Blob) bool {
	checksum, err := s.Checksum(blob)
	if err != nil {
		return false
	}

	return checksum == blob.Checksum
}

func (s *s3Store) NewBucketIterator(bucketName string) (BucketIterator, error) {
	s3Client := awss3.New(s.session)

	realBucketName := bucketName + "-" + s.identifier

	bucketExists, err := s.doesBucketExist(realBucketName)
	if err != nil {
		return nil, err
	}

	if !bucketExists {
		return nil, errors.New("bucket does not exist")
	}

	listObjectsOutput, err := s3Client.ListObjects(
		&awss3.ListObjectsInput{
			Bucket: aws.String(realBucketName),
		},
	)
	if err != nil {
		return nil, err
	}

	if len(listObjectsOutput.Contents) == 0 {
		return &s3BucketIterator{}, nil
	}

	doneCh := make(chan struct{})
	blobCh := make(chan *Blob)

	iterator := &s3BucketIterator{
		doneCh: doneCh,
		blobCh: blobCh,
	}

	go func() {
		for _, item := range listObjectsOutput.Contents {
			select {
			case <-doneCh:
				doneCh = nil
				return
			default:
				blobCh <- &Blob{
					Path: filepath.Join(
						bucketName,
						strings.TrimPrefix(*item.Key, realBucketName),
					),
				}
			}
		}

		close(blobCh)
	}()

	return iterator, nil
}
