package blobstore_test

import (
	. "github.com/c0-ops/goblob/blobstore"

	"code.cloudfoundry.org/lager"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	fakeboshsys "github.com/cloudfoundry/bosh-utils/system/fakes"
	"github.com/onsi/gomega/gbytes"
	"fmt"
)

var _ = Describe("LocalBlobstore", func() {
	var (
		outBuffer *gbytes.Buffer
		errBuffer *gbytes.Buffer
		logger    lager.Logger
		fs        *fakeboshsys.FakeFileSystem

		localBlobPath string

		localBlob LocalBlob
	)

	BeforeEach(func() {
		outBuffer = gbytes.NewBuffer()
		errBuffer = gbytes.NewBuffer()
		logger = lager.NewLogger("logger")
		logger.RegisterSink(lager.NewWriterSink(outBuffer, lager.INFO))
		logger.RegisterSink(lager.NewWriterSink(errBuffer, lager.ERROR))

		fs = fakeboshsys.NewFakeFileSystem()

		localBlobPath = "fake-local-blob-path"

		localBlob = NewLocalBlob(localBlobPath, fs, logger)
	})

	Describe("Path", func() {
		It("returns the local blob path", func() {
			Expect(localBlob.Path()).To(Equal(localBlobPath))
		})
	})

	Describe("Delete", func() {
		It("deletes the local blob from the file system", func() {
			err := fs.WriteFileString(localBlobPath, "fake-local-blob-content")
			Expect(err).ToNot(HaveOccurred())

			err = localBlob.Delete()
			Expect(err).ToNot(HaveOccurred())
			Expect(fs.FileExists(localBlobPath)).To(BeFalse())
		})

		Context("when deleting from the file system fails", func() {
			JustBeforeEach(func() {
				fs.RemoveAllStub = func(_ string) error {
					return bosherr.Error("fake-delete-error")
				}
			})

			It("returns an error", func() {
				err := localBlob.Delete()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-delete-error"))
			})
		})
	})

	Describe("DeleteSilently", func() {
		It("deletes the local blob from the file system", func() {
			err := fs.WriteFileString(localBlobPath, "fake-local-blob-content")
			Expect(err).ToNot(HaveOccurred())

			localBlob.DeleteSilently()
			Expect(fs.FileExists(localBlobPath)).To(BeFalse())
		})

		Context("when deleting from the file system fails", func() {
			JustBeforeEach(func() {
				fs.RemoveAllStub = func(_ string) error {
					return bosherr.Error("fake-delete-error")
				}
			})

			It("logs the error", func() {
				localBlob.DeleteSilently()

				Expect(errBuffer).To(gbytes.Say(
					fmt.Sprintf(`{"error":"Deleting local blob 'fake-local-blob-path': fake-delete-error","event":"Failed to delete local blob"}`),
				))
			})
		})
	})
})
