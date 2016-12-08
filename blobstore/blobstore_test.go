package blobstore_test

import (
	"errors"
	"io/ioutil"
	"strings"

	"code.cloudfoundry.org/lager"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	. "github.com/pivotal-customer0/cfblobmigrator/blobstore"

	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	fakeboshnfs "github.com/pivotal-customer0/cfblobmigrator/nfs/fakes"
)

var _ = Describe("Blobstore", func() {
	var (
		logBuffer     *gbytes.Buffer
		fakeNfsClient *fakeboshnfs.FakeClient
		fs            *fakesys.FakeFileSystem
		blobstore     Blobstore
	)

	BeforeEach(func() {
		fakeNfsClient = fakeboshnfs.NewFakeClient()
		fs = fakesys.NewFakeFileSystem()
		logger := lager.NewLogger("logger")
		logBuffer = gbytes.NewBuffer()

		blobstore = NewBlobstore(fakeNfsClient, fs, nil, logger)
	})

	Describe("Get", func() {
		BeforeEach(func() {
			fakeFile := fakesys.NewFakeFile("fake-destination-path", fs)
			fs.ReturnTempFile = fakeFile
		})

		It("gets the blob from the blobstore", func() {
			fakeNfsClient.GetContents = ioutil.NopCloser(strings.NewReader("fake-content"))
			file, err := fs.TempFile("bosh-init-local-blob")
			destinationPath := file.Name()
			err = file.Close()
			localBlob, err := blobstore.Get(destinationPath, "fake-blob-id")
			Expect(err).ToNot(HaveOccurred())
			defer localBlob.DeleteSilently()

			Expect(fakeNfsClient.GetPath).To(Equal("fake-blob-id"))
		})

		It("saves the blob to the destination path", func() {
			fakeNfsClient.GetContents = ioutil.NopCloser(strings.NewReader("fake-content"))

			file, err := fs.TempFile("bosh-init-local-blob")
			destinationPath := file.Name()
			err = file.Close()
			localBlob, err := blobstore.Get(destinationPath, "fake-blob-id")
			Expect(err).ToNot(HaveOccurred())
			defer func() {
				err := localBlob.Delete()
				Expect(err).ToNot(HaveOccurred())
			}()

			Expect(localBlob.Path()).To(Equal("fake-destination-path"))

			contents, err := fs.ReadFileString("fake-destination-path")
			Expect(err).ToNot(HaveOccurred())
			Expect(contents).To(Equal("fake-content"))
		})

		Context("when getting from blobstore fails", func() {
			It("returns an error", func() {
				fakeNfsClient.GetErr = errors.New("fake-get-error")

				_, err := blobstore.Get("", "fake-blob-id")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-get-error"))
			})
		})
	})

	Describe("Add", func() {
		BeforeEach(func() {
			fs.RegisterOpenFile("fake-source-path", &fakesys.FakeFile{
				Contents: []byte("fake-contents"),
			})
		})

		It("adds file to blobstore and returns its blob ID", func() {

			err := blobstore.Add("fake-source-path", "fake-blob-id")
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeNfsClient.PutPath).To(Equal("fake-blob-id"))
			Expect(fakeNfsClient.PutContents).To(Equal("fake-contents"))
		})
	})
})
