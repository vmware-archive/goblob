package s3_test

import (
	"bytes"
	"fmt"
	"io"
	"os"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	"github.com/c0-ops/goblob/s3"
	"github.com/minio/minio-go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client", func() {
	var (
		mc                                       *minio.Client
		bucketName                               string
		accessKeyID, secretAccessKey, s3Endpoint string
		region                                   string
		useSSL                                   bool
		config                                   s3.Config
		outBuffer                                *bytes.Buffer
		errBuffer                                *bytes.Buffer
		logger                                   boshlog.Logger
	)

	BeforeEach(func() {
		var err error
		bucketName = "mybucket"
		accessKeyID = "AKIAIOSFODNN7EXAMPLE"
		secretAccessKey = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
		region = "us-east-1"
		useSSL = false

		if os.Getenv("MINIO_PORT_9000_TCP_ADDR") == "" {
			s3Endpoint = "127.0.0.1:9000"
		} else {
			s3Endpoint = fmt.Sprintf("%s:9000", os.Getenv("MINIO_PORT_9000_TCP_ADDR"))
		}

		outBuffer = bytes.NewBufferString("")
		errBuffer = bytes.NewBufferString("")
		logger = boshlog.NewWriterLogger(boshlog.LevelDebug, outBuffer, errBuffer)

		mc, err = minio.New(s3Endpoint, accessKeyID, secretAccessKey, false)
		Expect(err).NotTo(HaveOccurred())
	})

	JustBeforeEach(func() {
		config = s3.Config{
			Endpoint:        s3Endpoint,
			AccessKeyID:     accessKeyID,
			SecretAccessKey: secretAccessKey,
			Region:          region,
			UseSSL:          useSSL,
		}
	})

	Describe("CreateBucket", func() {
		Context("when the bucket already exists", func() {
			var (
				createErr error
			)

			JustBeforeEach(func() {
				client, err := s3.NewClient(config, logger)
				Expect(err).NotTo(HaveOccurred())
				err = mc.MakeBucket(bucketName, region)
				if err != nil {
					_, err := mc.BucketExists(bucketName)
					Expect(err).NotTo(HaveOccurred())
				}
				createErr = client.CreateBucket(bucketName)
			})

			AfterEach(func() {
				err := mc.RemoveBucket(bucketName)
				Expect(err).NotTo(HaveOccurred())
			})

			It("does not return an error", func() {
				Expect(createErr).NotTo(HaveOccurred())
			})

			It("provides logging", func() {
				outString := outBuffer.String()
				Expect(outString).To(ContainSubstring("Start creating bucket"))
				Expect(outString).To(ContainSubstring("Bucket already exists"))
				Expect(outString).To(ContainSubstring("Done creating bucket"))
			})
		})

		Context("when the bucket does not exist", func() {
			var (
				createErr error
			)

			JustBeforeEach(func() {
				client, err := s3.NewClient(config, logger)
				Expect(err).NotTo(HaveOccurred())
				bucketList, err := mc.ListBuckets()
				Expect(err).NotTo(HaveOccurred())
				Expect(bucketList).To(HaveLen(1))

				createErr = client.CreateBucket(bucketName)
			})

			AfterEach(func() {
				err := mc.RemoveBucket(bucketName)
				Expect(err).NotTo(HaveOccurred())
			})

			It("does not return an error", func() {
				Expect(createErr).NotTo(HaveOccurred())
			})

			It("provides logging", func() {
				outString := outBuffer.String()
				Expect(outString).To(ContainSubstring("Start creating bucket"))
				Expect(outString).To(ContainSubstring("Done creating bucket"))
			})

			It("creates the bucket", func() {
				bucketList, err := mc.ListBuckets()
				Expect(err).NotTo(HaveOccurred())
				Expect(bucketList).To(HaveLen(2))
				Expect(bucketList[1].Name).To(Equal(bucketName))
			})
		})

		Context("when the region does not exist", func() {
			var (
				createErr error
			)

			BeforeEach(func() {
				region = "fake-region"
			})

			JustBeforeEach(func() {
				client, err := s3.NewClient(config, logger)
				Expect(err).NotTo(HaveOccurred())
				createErr = client.CreateBucket(bucketName)
			})

			It("return an error", func() {
				Expect(createErr).To(HaveOccurred())
			})

			It("provides logging", func() {
				outString := outBuffer.String()
				Expect(outString).To(ContainSubstring("Start creating bucket"))

				errorLogString := errBuffer.String()
				Expect(errorLogString).To(ContainSubstring("Failed to create bucket"))
			})
		})
	})

	Describe("UploadObject", func() {
		var (
			objectName  string
			object      io.ReadCloser
			filePath    string
			contentType string
			size        int64
			uploadErr   error
		)

		Context("when the object exists on the file system", func() {
			JustBeforeEach(func() {
				client, err := s3.NewClient(config, logger)
				Expect(err).NotTo(HaveOccurred())
				err = mc.MakeBucket(bucketName, region)
				if err != nil {
					_, err := mc.BucketExists(bucketName)
					Expect(err).NotTo(HaveOccurred())
				}
				err = client.CreateBucket(bucketName)
				Expect(err).NotTo(HaveOccurred())

				objectName = "test"
				filePath = "fixtures/test.txt"
				contentType = "text/plain"
				object, err = os.Open(filePath)
				Expect(err).NotTo(HaveOccurred())
				defer object.Close()

				size, uploadErr = client.UploadObject(bucketName, objectName, object, contentType)
			})

			AfterEach(func() {
				err := mc.RemoveObject(bucketName, objectName)
				Expect(err).NotTo(HaveOccurred())
				err = mc.RemoveBucket(bucketName)
				Expect(err).NotTo(HaveOccurred())
			})

			It("does not return an error", func() {
				Expect(uploadErr).NotTo(HaveOccurred())
			})

			It("file size is larger than 0", func() {
				Expect(size).Should(BeNumerically(">", 0))
			})

			It("provides logging", func() {
				outString := outBuffer.String()
				Expect(outString).To(ContainSubstring("Start creating bucket"))
				Expect(outString).To(ContainSubstring("Bucket already exists"))
				Expect(outString).To(ContainSubstring("Done creating bucket"))
				Expect(outString).To(ContainSubstring("Start uploading object"))
				Expect(outString).To(ContainSubstring("Done uploading object"))
			})
		})

		Context("when the object does not exist on the file system", func() {
			JustBeforeEach(func() {
				client, err := s3.NewClient(config, logger)
				Expect(err).NotTo(HaveOccurred())
				err = mc.MakeBucket(bucketName, region)
				if err != nil {
					_, err := mc.BucketExists(bucketName)
					Expect(err).NotTo(HaveOccurred())
				}
				err = client.CreateBucket(bucketName)
				Expect(err).NotTo(HaveOccurred())

				objectName = ""
				filePath = "fixtures/test.txt"
				contentType = "text/plain"
				object, err = os.Open(filePath)
				Expect(err).NotTo(HaveOccurred())
				defer object.Close()

				size, uploadErr = client.UploadObject(bucketName, objectName, object, contentType)
			})

			AfterEach(func() {
				err := mc.RemoveBucket(bucketName)
				Expect(err).NotTo(HaveOccurred())
			})

			It("return an error", func() {
				Expect(uploadErr).To(HaveOccurred())
			})

			It("provides logging", func() {
				outString := outBuffer.String()
				Expect(outString).To(ContainSubstring("Start creating bucket"))
				Expect(outString).To(ContainSubstring("Bucket already exists"))
				Expect(outString).To(ContainSubstring("Done creating bucket"))
				Expect(outString).To(ContainSubstring("Start uploading object"))

				errorLogString := errBuffer.String()
				Expect(errorLogString).To(ContainSubstring("Failed to upload object Object name cannot be empty"))
			})
		})
	})
})
