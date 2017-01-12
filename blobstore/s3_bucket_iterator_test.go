package blobstore_test

import (
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	awss3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/c0-ops/goblob/blobstore"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("S3BucketIterator", func() {
	var (
		iterator                 blobstore.BucketIterator
		store                    blobstore.Blobstore
		bucketName               string
		bucketNameWithIdentifier string
		s3Client                 *awss3.S3
	)

	const (
		s3Region = "us-east-1"
	)

	BeforeEach(func() {
		var s3Endpoint string

		if os.Getenv("MINIO_PORT_9000_TCP_ADDR") == "" {
			s3Endpoint = "http://127.0.0.1:9000"
		} else {
			s3Endpoint = fmt.Sprintf("http://%s:9000", os.Getenv("MINIO_PORT_9000_TCP_ADDR"))
		}

		session := session.New(&aws.Config{
			Region: aws.String(s3Region),
			Credentials: credentials.NewStaticCredentials(
				"AKIAIOSFODNN7EXAMPLE",
				"wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				"example-token",
			),
			Endpoint:         aws.String(s3Endpoint),
			DisableSSL:       aws.Bool(true),
			S3ForcePathStyle: aws.Bool(true),
		})

		s3Client = awss3.New(session)

		bucketName = fmt.Sprintf("some-bucket-%d", GinkgoParallelNode())
		bucketNameWithIdentifier = fmt.Sprintf("%s-%s", bucketName, "identifier")

		_, err := s3Client.CreateBucket(&awss3.CreateBucketInput{
			Bucket: aws.String(bucketNameWithIdentifier),
		})
		Expect(err).NotTo(HaveOccurred())

		store = blobstore.NewS3(
			"identifier",
			"AKIAIOSFODNN7EXAMPLE",
			"wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			s3Region,
			s3Endpoint,
			true,
		)

		iterator, err = store.NewBucketIterator(bucketName)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		listObjectsOutput, err := s3Client.ListObjects(&awss3.ListObjectsInput{
			Bucket: aws.String(bucketNameWithIdentifier),
		})
		Expect(err).NotTo(HaveOccurred())

		for _, item := range listObjectsOutput.Contents {
			_, err := s3Client.DeleteObject(&awss3.DeleteObjectInput{
				Bucket: aws.String(bucketNameWithIdentifier),
				Key:    item.Key,
			})
			Expect(err).NotTo(HaveOccurred())
		}

		_, err = s3Client.DeleteBucket(&awss3.DeleteBucketInput{
			Bucket: aws.String(bucketNameWithIdentifier),
		})
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Next", func() {
		It("returns an error", func() {
			_, err := iterator.Next()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("no more items in iterator"))
		})

		Context("when a blob exists in the bucket", func() {
			var expectedBlob blobstore.Blob

			BeforeEach(func() {
				expectedBlob = blobstore.Blob{
					Path: fmt.Sprintf("%s/some-path/some-file", bucketName),
				}

				_, err := s3Client.PutObject(&awss3.PutObjectInput{
					Body:   strings.NewReader("content"),
					Bucket: aws.String(bucketNameWithIdentifier),
					Key:    aws.String("some-path/some-file"),
					Metadata: map[string]*string{
						"Checksum": aws.String("some-checksum"),
					},
				})
				Expect(err).NotTo(HaveOccurred())

				iterator, err = store.NewBucketIterator(bucketName)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns the blob", func() {
				blob, err := iterator.Next()
				Expect(err).NotTo(HaveOccurred())
				Expect(*blob).To(Equal(expectedBlob))
			})

			It("returns an error when all blobs have been listed", func() {
				_, err := iterator.Next()
				Expect(err).NotTo(HaveOccurred())

				_, err = iterator.Next()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("no more items in iterator"))
			})
		})
	})

	Describe("Done", func() {
		Context("when blobs exist in the bucket", func() {
			BeforeEach(func() {
				_, err := s3Client.PutObject(&awss3.PutObjectInput{
					Body:   strings.NewReader("content"),
					Bucket: aws.String(bucketNameWithIdentifier),
					Key:    aws.String("some-path/some-file"),
					Metadata: map[string]*string{
						"Checksum": aws.String("some-checksum"),
					},
				})
				Expect(err).NotTo(HaveOccurred())

				_, err = s3Client.PutObject(&awss3.PutObjectInput{
					Body:   strings.NewReader("content"),
					Bucket: aws.String(bucketNameWithIdentifier),
					Key:    aws.String("some-path/some-other-file"),
					Metadata: map[string]*string{
						"Checksum": aws.String("some-checksum"),
					},
				})
				Expect(err).NotTo(HaveOccurred())

				iterator, err = store.NewBucketIterator(bucketName)
				Expect(err).NotTo(HaveOccurred())
			})

			It("causes Next to return an error", func() {
				iterator.Done()

				_, err := iterator.Next()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("no more items in iterator"))
			})
		})
	})
})
