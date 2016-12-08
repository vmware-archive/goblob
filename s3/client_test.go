package s3_test

import (
	"bytes"
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
		mc                           *minio.Client
		mcErr                        error
		client                       s3.Client
		bucketName                   string
		region                       string
		accessKeyID, secretAccessKey string
		outBuffer                    *bytes.Buffer
		errBuffer                    *bytes.Buffer
		logger                       boshlog.Logger
	)

	BeforeEach(func() {
		bucketName = "mybucket"
		region = "us-east-1"
		accessKeyID = "D2Z5WU2UI35D05WXSJGW"
		secretAccessKey = "Y+4XHK07GQbDqQbkVFIgz2VVi3fapWIGfsdpIL0q"

		outBuffer = bytes.NewBufferString("")
		errBuffer = bytes.NewBufferString("")
		logger = boshlog.NewWriterLogger(boshlog.LevelDebug, outBuffer, errBuffer)

		mc, mcErr = minio.New(fakeS3EndpointURL, accessKeyID, secretAccessKey, false)
		Expect(mcErr).NotTo(HaveOccurred())
		client, _ = s3.NewClient(fakeS3EndpointURL, accessKeyID, secretAccessKey, false, logger)
	})

	Describe("CreateBucket", func() {
		Context("when the bucket already exists", func() {
			var (
				createErr error
			)

			BeforeEach(func() {
				err := mc.MakeBucket(bucketName, region)
				if err != nil {
					_, err := mc.BucketExists(bucketName)
					Expect(err).NotTo(HaveOccurred())
				}
				createErr = client.CreateBucket(bucketName, region)
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

			BeforeEach(func() {
				bucketList, err := mc.ListBuckets()
				Expect(err).NotTo(HaveOccurred())
				Expect(bucketList).To(HaveLen(0))

				createErr = client.CreateBucket(bucketName, region)
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
				Expect(bucketList).To(HaveLen(1))
				Expect(bucketList[0].Name).To(Equal(bucketName))
			})
		})

		Context("when the region does not exist", func() {
			var (
				createErr   error
				bogusRegion = "fake-region"
			)

			BeforeEach(func() {
				createErr = client.CreateBucket(bucketName, bogusRegion)
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

		Context("when passed invalid parameters", func() {
			var (
				err error
			)

			BeforeEach(func() {
				client, err = s3.NewClient("http://fake-endpoint", "fake-access-id", "fake-secret-key", false, logger)
			})

			It("returns a nil client and an error", func() {
				Expect(client).To(BeNil())
				Expect(err).To(HaveOccurred())
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
			BeforeEach(func() {
				err := mc.MakeBucket(bucketName, region)
				if err != nil {
					_, err := mc.BucketExists(bucketName)
					Expect(err).NotTo(HaveOccurred())
				}
				err = client.CreateBucket(bucketName, region)
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
			BeforeEach(func() {
				err := mc.MakeBucket(bucketName, region)
				if err != nil {
					_, err := mc.BucketExists(bucketName)
					Expect(err).NotTo(HaveOccurred())
				}
				err = client.CreateBucket(bucketName, region)
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
