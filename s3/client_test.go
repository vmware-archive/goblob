package s3_test

import (
	"fmt"
	"io"
	"os"

	"code.cloudfoundry.org/lager"
	"github.com/minio/minio-go"
	"github.com/c0-ops/goblob/s3"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Client", func() {
	var (
		mc                           *minio.Client
		mcErr                        error
		client                       s3.Client
		bucketName                   string
		region                       string
		accessKeyID, secretAccessKey string
		logBuffer                    *gbytes.Buffer
		logger                       lager.Logger
	)

	BeforeEach(func() {
		bucketName = "mybucket"
		region = "us-east-1"
		accessKeyID = "JP0NF5645O4O6A67VGA8"
		secretAccessKey = "EeaNuDd5zcNLN7WV4+x50TzYyKckHg/R1FwwPbbo"

		logger = lager.NewLogger("logger")
		logBuffer = gbytes.NewBuffer()
		logger.RegisterSink(lager.NewWriterSink(logBuffer, lager.INFO))

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
				Expect(logBuffer).To(gbytes.Say(
					fmt.Sprintf(`{"bucket_name":"%s","event":"starting","region":"%s"}`, bucketName, region),
				))
				Expect(logBuffer).To(gbytes.Say(
					fmt.Sprintf(`{"bucket_name":"%s","event":"already-exists","region":"%s"}`, bucketName, region),
				))
				Expect(logBuffer).To(gbytes.Say(
					fmt.Sprintf(`{"bucket_name":"%s","event":"done","region":"%s"}`, bucketName, region),
				))
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
				Expect(logBuffer).To(gbytes.Say(
					fmt.Sprintf(`{"bucket_name":"%s","event":"starting","region":"%s"}`, bucketName, region),
				))
				Expect(logBuffer).To(gbytes.Say(
					fmt.Sprintf(`{"bucket_name":"%s","event":"done","region":"%s"}`, bucketName, region),
				))
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
				Expect(logBuffer).To(gbytes.Say(
					fmt.Sprintf(`{"bucket_name":"%s","event":"starting","region":"%s"}`, bucketName, bogusRegion),
				))
				Expect(logBuffer).To(gbytes.Say(
					fmt.Sprintf(`{"bucket_name":"%s","error":"%s","event":"failed","region":"%s"}`, bucketName, createErr, bogusRegion),
				))
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
				Expect(logBuffer).To(gbytes.Say(
					fmt.Sprintf(`{"bucket_name":"%s","event":"starting","region":"%s"}`, bucketName, region),
				))
				Expect(logBuffer).To(gbytes.Say(
					fmt.Sprintf(`{"bucket_name":"%s","event":"already-exists","region":"%s"}`, bucketName, region),
				))
				Expect(logBuffer).To(gbytes.Say(
					fmt.Sprintf(`{"bucket_name":"%s","event":"done","region":"%s"}`, bucketName, region),
				))

				Expect(logBuffer).To(gbytes.Say(
					fmt.Sprintf(`{"bucket_name":"%s","content_type":"%s","event":"uploading","object_name":"%s","size":%d}`, bucketName, contentType, objectName, 0),
				))
				Expect(logBuffer).To(gbytes.Say(
					fmt.Sprintf(`{"bucket_name":"%s","content_type":"%s","event":"done","object_name":"%s","size":%d}`, bucketName, contentType, objectName, size),
				))
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
				Expect(logBuffer).To(gbytes.Say(
					fmt.Sprintf(`{"bucket_name":"%s","event":"starting","region":"%s"}`, bucketName, region),
				))
				Expect(logBuffer).To(gbytes.Say(
					fmt.Sprintf(`{"bucket_name":"%s","event":"already-exists","region":"%s"}`, bucketName, region),
				))
				Expect(logBuffer).To(gbytes.Say(
					fmt.Sprintf(`{"bucket_name":"%s","event":"done","region":"%s"}`, bucketName, region),
				))

				Expect(logBuffer).To(gbytes.Say(
					fmt.Sprintf(`{"bucket_name":"%s","content_type":"%s","event":"uploading","object_name":"%s","size":%d}`, bucketName, contentType, objectName, 0),
				))
				Expect(logBuffer).To(gbytes.Say(
					fmt.Sprintf(`{"bucket_name":"%s","content_type":"%s","error":"%s","event":"failed","object_name":"%s","size":%d}`, bucketName, contentType, uploadErr.Error(), objectName, 0),
				))
			})
		})
	})
})
