package blobstore_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/c0-ops/goblob/blobstore"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"

	"github.com/aws/aws-sdk-go/aws/session"

	awss3 "github.com/aws/aws-sdk-go/service/s3"
)

var _ = Describe("S3Store", func() {
	var s3Endpoint string
	region := "us-east-1"
	accessKey := "AKIAIOSFODNN7EXAMPLE"
	secretKey := "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
	if os.Getenv("MINIO_PORT_9000_TCP_ADDR") == "" {
		s3Endpoint = "http://127.0.0.1:9000"
	} else {
		s3Endpoint = fmt.Sprintf("http://%s:9000", os.Getenv("MINIO_PORT_9000_TCP_ADDR"))
	}
	config := &aws.Config{
		Region:           aws.String(region),
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Endpoint:         aws.String(s3Endpoint),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
	}
	controlBucket := "cc-buildpacks-identifier"

	testsToRun("Multi-part", config, controlBucket, blobstore.NewS3("identifier", accessKey, secretKey, region, s3Endpoint, true))
	testsToRun("non Multi-part", config, controlBucket, blobstore.NewS3("identifier", accessKey, secretKey, region, s3Endpoint, false))
})

func testsToRun(testSuiteName string, config *aws.Config, controlBucket string, store blobstore.Blobstore) {
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
	Describe(testSuiteName, func() {
		Describe("List()", func() {
			It("Should return list of files", func() {
				fileReader, err := os.Open("./s3_testdata/test.txt")
				Ω(err).ShouldNot(HaveOccurred())
				for _, path := range []string{"cc-buildpacks/aa/bb", "cc-buildpacks/aa/cc", "cc-buildpacks/aa/dd"} {
					err := store.Write(&blobstore.Blob{
						Path: filepath.Join(path, "test.txt"),
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
				fileReader, err := os.Open("./s3_testdata/test.txt")
				Ω(err).ShouldNot(HaveOccurred())
				writeErr := store.Write(&blobstore.Blob{
					Path: "cc-buildpacks/aa/bb/test.txt",
				}, fileReader)
				Ω(writeErr).ShouldNot(HaveOccurred())
				reader, err := store.Read(&blobstore.Blob{
					Path: "cc-buildpacks/aa/bb/test.txt",
				})
				Ω(err).ShouldNot(HaveOccurred())
				Ω(reader).ShouldNot(BeNil())
			})
		})
		Describe("Write()", func() {
			It("Should write to s3 blob store with correct checksum", func() {
				reader, err := os.Open("./s3_testdata/test.txt")
				Ω(err).ShouldNot(HaveOccurred())
				blob := &blobstore.Blob{
					Path: "cc-buildpacks/aa/bb/test.txt",
				}
				err = store.Write(blob, reader)
				Ω(err).ShouldNot(HaveOccurred())

				checksum, err := store.Checksum(blob)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(checksum).Should(BeEquivalentTo("d8e8fca2dc0f896fd7cb4cb0031ba249"))
			})
		})
	})
}
