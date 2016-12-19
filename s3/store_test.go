package s3_test

import (
	"errors"
	"fmt"
	"os"

	. "github.com/c0-ops/goblob/s3"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/c0-ops/goblob"

	awss3 "github.com/aws/aws-sdk-go/service/s3"
)

var _ = Describe("S3Store", func() {
	var store goblob.Store
	var s3Endpoint string
	var config *aws.Config
	var controlBucket string
	BeforeEach(func() {
		if os.Getenv("MINIO_PORT_9000_TCP_ADDR") == "" {
			s3Endpoint = "http://127.0.0.1:9000"
		} else {
			s3Endpoint = fmt.Sprintf("http://%s:9000", os.Getenv("MINIO_PORT_9000_TCP_ADDR"))
		}
		config = &aws.Config{
			Region:           aws.String("us-east-1"),
			Credentials:      credentials.NewStaticCredentials("AKIAIOSFODNN7EXAMPLE", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", ""),
			Endpoint:         aws.String(s3Endpoint),
			DisableSSL:       aws.Bool(true),
			S3ForcePathStyle: aws.Bool(true),
		}
		store = New(config)
		controlBucket = "cc-buildpackets-identifier"
	})

	AfterEach(func() {
		session := session.New(config)
		s3Service := awss3.New(session)
		listObjectsOutput, err := s3Service.ListObjects(&awss3.ListObjectsInput{
			Bucket: aws.String(controlBucket),
		})

		if err == nil {
			for _, item := range listObjectsOutput.Contents {
				s3Service.DeleteObject(&awss3.DeleteObjectInput{
					Bucket: aws.String(controlBucket),
					Key:    item.Key,
				})
			}
			s3Service.DeleteBucket(&awss3.DeleteBucketInput{
				Bucket: aws.String(controlBucket),
			})
		}
	})

	Describe("List()", func() {
		It("Should return an error", func() {
			err := errors.New("not implemented")
			_, listErr := store.List()
			Ω(listErr).Should(BeEquivalentTo(err))
		})
	})
	Describe("Read()", func() {
		It("Should read the file", func() {
			fileReader, err := os.Open("./fixtures/test.txt")
			Ω(err).ShouldNot(HaveOccurred())
			writeErr := store.Write(goblob.Blob{
				Path:     controlBucket + "/aa/bb",
				Filename: "test.txt",
				Checksum: "d8e8fca2dc0f896fd7cb4cb0031ba249",
			}, fileReader)
			Ω(writeErr).ShouldNot(HaveOccurred())
			reader, err := store.Read(goblob.Blob{
				Path:     controlBucket + "/aa/bb",
				Filename: "test.txt",
				Checksum: "d8e8fca2dc0f896fd7cb4cb0031ba249"})
			Ω(err).ShouldNot(HaveOccurred())
			Ω(reader).ShouldNot(BeNil())
		})
	})
	Describe("Write()", func() {
		It("Should write to s3 blob store", func() {
			reader, err := os.Open("./fixtures/test.txt")
			Ω(err).ShouldNot(HaveOccurred())
			writeErr := store.Write(goblob.Blob{
				Path:     controlBucket + "/aa/bb",
				Filename: "test.txt",
				Checksum: "d8e8fca2dc0f896fd7cb4cb0031ba249",
			}, reader)
			Ω(writeErr).ShouldNot(HaveOccurred())
		})
	})
})
