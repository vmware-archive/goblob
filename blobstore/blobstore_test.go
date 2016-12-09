package blobstore_test

import (
	"errors"
	"io/ioutil"
	"strings"

	. "github.com/c0-ops/goblob/blobstore"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/c0-ops/goblob/nfs/fakes"
	faketar "github.com/c0-ops/goblob/tar/fakes"
)

var _ = Describe("Blobstore", func() {
	var (
		fakeNfsClient *fakes.FakeClient
		fs            *fakesys.FakeFileSystem
		fakeRunner    *fakesys.FakeCmdRunner
		extractor     faketar.FakeCmdExtractor
		blobstore     Blobstore
	)

	BeforeEach(func() {
		fakeNfsClient = fakes.NewFakeClient()
		fakeRunner = fakesys.NewFakeCmdRunner()
		fs = fakesys.NewFakeFileSystem()
		logger := boshlog.NewLogger(boshlog.LevelNone)
		extractor = faketar.NewFakeCmdExtractor()

		blobstore = NewBlobstore(fakeNfsClient, fs, extractor, logger)
	})

	Describe("Get", func() {
		BeforeEach(func() {
			fakeFile := fakesys.NewFakeFile("fake-destination-path", fs)
			fs.ReturnTempFile = fakeFile
		})

		It("gets the blob from the blobstore", func() {
			fakeNfsClient.GetContents = ioutil.NopCloser(strings.NewReader("fake-content"))

			localBlob, err := blobstore.Get("fake-destination-path", "fake-blob-id")
			Expect(err).ToNot(HaveOccurred())
			defer localBlob.DeleteSilently()

			Expect(fakeNfsClient.GetPath).To(Equal("fake-destination-path"))
		})

		It("saves the blob to the destination path", func() {
			fakeNfsClient.GetContents = ioutil.NopCloser(strings.NewReader("fake-content"))

			localBlob, err := blobstore.Get("fake-destination-path", "fake-blob-id")
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

				_, err := blobstore.Get("path/to/blobstore", "fake-blob-id")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-get-error"))
			})
		})
	})

	Describe("Add", func() {
		BeforeEach(func() {
			fs.RegisterOpenFile("fake-source-path", &fakesys.FakeFile{
				Contents: []byte("fake-content"),
			})
		})

		It("adds file to blobstore and returns its blob ID", func() {

			err := blobstore.Add("fake-source-path", "fake-blob-id")
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeNfsClient.PutPath).To(Equal("fake-source-path"))
			Expect(fakeNfsClient.PutContents).To(Equal("fake-content"))
		})
	})
})
