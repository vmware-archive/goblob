package blobstore_test

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/Azure/azure-storage-blob-go/2018-03-28/azblob"
	"github.com/pivotal-cf/goblob/blobstore"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("azblobStore", func() {
	var (
		cloudStorageEnpointsMap = map[string]string{
			"AzureChinaCloud":   "core.chinacloudapi.cn",
			"AzureCloud":        "core.windows.net",
			"AzureGermanCloud":  "core.cloudapi.de",
			"AzureUSGovernment": "core.usgovcloudapi.net",
		}
	)

	accountName := os.Getenv("AZURE_STORAGE_ACCOUNT")
	accountKey := os.Getenv("AZURE_STORAGE_ACCOUNT_KEY")
	cloudName := os.Getenv("AZURE_CLOUD")
	if os.Getenv("AZURE_CLOUD") == "" {
		cloudName = "AzureCloud"
	}
	controlContainer := "some-buildpacks"

	blobStore := blobstore.NewAzBlobStore(accountName, accountKey, cloudName, "some-buildpacks", "some-droplets", "some-packages", "some-resources")

	credential := azblob.NewSharedKeyCredential(accountName, accountKey)
	pipeline := azblob.NewPipeline(credential, azblob.PipelineOptions{})
	primaryURL, _ := url.Parse(
		fmt.Sprintf("https://%s.blob.%s", accountName, cloudStorageEnpointsMap[cloudName]))
	serviceURL := azblob.NewServiceURL(*primaryURL, pipeline)

	AfterSuite(func() {
		containerURL := serviceURL.NewContainerURL(controlContainer)
		_, _ = containerURL.Delete(context.Background(), azblob.ContainerAccessConditions{})
	})

	Describe("Name()", func() {
		It("Should return the name", func() {
			name := blobStore.Name()
			Expect(name).Should(BeEquivalentTo("azure blob store"))
		})
	})

	Describe("List()", func() {
		It("Should return list of files", func() {
			fileReader, err := os.Open("./s3_testdata/test.txt")
			Expect(err).ShouldNot(HaveOccurred())
			for _, path := range []string{"cc-buildpacks/aa/bb", "cc-buildpacks/aa/cc", "cc-buildpacks/aa/dd"} {
				err := blobStore.Write(&blobstore.Blob{
					Path: filepath.Join(path, "test.txt"),
				}, fileReader)
				Expect(err).ShouldNot(HaveOccurred())
			}
			blobs, err := blobStore.List()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(blobs)).Should(BeEquivalentTo(3))
		})
	})

	Describe("Read()", func() {
		It("Should read the file", func() {
			fileReader, err := os.Open("./s3_testdata/test.txt")
			Expect(err).ShouldNot(HaveOccurred())
			writeErr := blobStore.Write(&blobstore.Blob{
				Path: "cc-buildpacks/aa/bb/test.txt",
			}, fileReader)
			Expect(writeErr).ShouldNot(HaveOccurred())
			reader, err := blobStore.Read(&blobstore.Blob{
				Path: "cc-buildpacks/aa/bb/test.txt",
			})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(reader).ShouldNot(BeNil())
		})
	})

	Describe("Write()", func() {
		It("Should write to azure blob store with correct checksum", func() {
			reader, err := os.Open("./s3_testdata/test.txt")
			Expect(err).ShouldNot(HaveOccurred())
			blob := &blobstore.Blob{
				Path: "cc-buildpacks/aa/bb/test.txt",
			}
			err = blobStore.Write(blob, reader)
			Expect(err).ShouldNot(HaveOccurred())

			checksum, err := blobStore.Checksum(blob)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(checksum).Should(BeEquivalentTo("d8e8fca2dc0f896fd7cb4cb0031ba249"))
		})
	})
})
