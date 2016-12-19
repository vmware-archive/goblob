package s3_test

import (
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
	var cleanup bool
	BeforeEach(func() {
		region := "us-east-1"
		accessKey := "AKIAIOSFODNN7EXAMPLE"
		secretKey := "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
		if os.Getenv("MINIO_PORT_9000_TCP_ADDR") == "" {
			s3Endpoint = "http://127.0.0.1:9000"
		} else {
			s3Endpoint = fmt.Sprintf("http://%s:9000", os.Getenv("MINIO_PORT_9000_TCP_ADDR"))
		}
		config = &aws.Config{
			Region:           aws.String(region),
			Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
			Endpoint:         aws.String(s3Endpoint),
			DisableSSL:       aws.Bool(true),
			S3ForcePathStyle: aws.Bool(true),
		}
		store = New("identifier", accessKey, secretKey, region, s3Endpoint)
		controlBucket = "cc-buildpackets-identifier"
		cleanup = true
	})

	AfterEach(func() {
		if cleanup {
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
		}
	})

	Describe("List()", func() {
		It("Should return list of files", func() {
			fileReader, err := os.Open("./fixtures/test.txt")
			Ω(err).ShouldNot(HaveOccurred())
			for _, path := range []string{"cc-buildpacks/aa/bb", "cc-buildpacks/aa/cc", "cc-buildpacks/aa/dd"} {
				err := store.Write(&goblob.Blob{
					Path:     path,
					Filename: "test.txt",
					Checksum: "d8e8fca2dc0f896fd7cb4cb0031ba249",
				}, fileReader)
				Ω(err).ShouldNot(HaveOccurred())
			}
			blobs, err := store.List()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(len(blobs)).Should(BeEquivalentTo(3))
		})
	})
	Describe("Read()", func() {
		It("Should read the file", func() {
			fileReader, err := os.Open("./fixtures/test.txt")
			Ω(err).ShouldNot(HaveOccurred())
			writeErr := store.Write(&goblob.Blob{
				Path:     "cc-buildpackets/aa/bb",
				Filename: "test.txt",
				Checksum: "d8e8fca2dc0f896fd7cb4cb0031ba249",
			}, fileReader)
			Ω(writeErr).ShouldNot(HaveOccurred())
			reader, err := store.Read(&goblob.Blob{
				Path:     "cc-buildpackets/aa/bb",
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
			writeErr := store.Write(&goblob.Blob{
				Path:     "cc-buildpackets/aa/bb",
				Filename: "test.txt",
				Checksum: "d8e8fca2dc0f896fd7cb4cb0031ba249",
			}, reader)
			Ω(writeErr).ShouldNot(HaveOccurred())
		})
	})
})
