package goblob_test

import (
	"errors"

	"github.com/c0-ops/goblob"
	"github.com/c0-ops/goblob/goblobfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BlobstoreMigrator", func() {
	var (
		migrator     goblob.BlobstoreMigrator
		blobMigrator *goblobfakes.FakeBlobMigrator
		dstStore     *goblobfakes.FakeBlobstore
		srcStore     *goblobfakes.FakeBlobstore
	)

	BeforeEach(func() {
		dstStore = &goblobfakes.FakeBlobstore{}
		srcStore = &goblobfakes.FakeBlobstore{}
		blobMigrator = &goblobfakes.FakeBlobMigrator{}
		migrator = goblob.NewBlobstoreMigrator(1, blobMigrator)
	})

	Describe("Migrate", func() {
		var firstBlob, secondBlob, thirdBlob *goblob.Blob

		BeforeEach(func() {
			firstBlob = &goblob.Blob{
				Checksum: "some-file-checksum",
				Path:     "some-file-path/some-file",
			}

			secondBlob = &goblob.Blob{
				Checksum: "some-other-file-checksum",
				Path:     "some-other-path/some-other-file",
			}

			thirdBlob = &goblob.Blob{
				Checksum: "yet-another-file-checksum",
				Path:     "yet-another-path/yet-another-file",
			}

			srcStore.ListReturns([]*goblob.Blob{firstBlob, secondBlob, thirdBlob}, nil)
		})

		It("uploads all the files from the source", func() {
			err := migrator.Migrate(dstStore, srcStore)
			Expect(err).NotTo(HaveOccurred())
			Expect(blobMigrator.MigrateCallCount()).To(Equal(3))
			Expect(blobMigrator.MigrateArgsForCall(0)).To(Equal(firstBlob))
			Expect(blobMigrator.MigrateArgsForCall(1)).To(Equal(secondBlob))
			Expect(blobMigrator.MigrateArgsForCall(2)).To(Equal(thirdBlob))
		})

		Context("when a file already exists", func() {
			BeforeEach(func() {
				dstStore.ExistsStub = func(blob *goblob.Blob) bool {
					return blob.Path == "some-other-path/some-other-file"
				}
			})

			It("uploads only the new files", func() {
				err := migrator.Migrate(dstStore, srcStore)
				Expect(err).NotTo(HaveOccurred())
				Expect(blobMigrator.MigrateCallCount()).To(Equal(2))
				Expect(blobMigrator.MigrateArgsForCall(0)).To(Equal(firstBlob))
				Expect(blobMigrator.MigrateArgsForCall(1)).To(Equal(thirdBlob))
			})
		})

		Context("when there is an error uploading one blob", func() {
			BeforeEach(func() {
				blobMigrator.MigrateStub = func(blob *goblob.Blob) error {
					if blob.Path == "some-other-path/some-other-file" {
						return errors.New("migrate-err")
					}
					return nil
				}
			})

			It("stops uploading", func() {
				migrator.Migrate(dstStore, srcStore)
				Expect(blobMigrator.MigrateCallCount()).To(Equal(2))
				Expect(blobMigrator.MigrateArgsForCall(0)).To(Equal(firstBlob))
				Expect(blobMigrator.MigrateArgsForCall(1)).To(Equal(secondBlob))
			})

			It("returns an error", func() {
				err := migrator.Migrate(dstStore, srcStore)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("migrate-err"))
			})
		})

		It("returns an error when the source store is nil", func() {
			err := migrator.Migrate(dstStore, nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("src is an empty store"))
		})

		It("returns an error when the destination store is nil", func() {
			err := migrator.Migrate(nil, srcStore)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("dst is an empty store"))
		})

		Context("when the source store has no files", func() {
			BeforeEach(func() {
				srcStore.ListReturns(nil, nil)
			})

			It("returns an error", func() {
				err := migrator.Migrate(dstStore, srcStore)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("the source store has no files"))
			})
		})

		Context("when there is an error listing the source's files", func() {
			BeforeEach(func() {
				srcStore.ListReturns(nil, errors.New("list-error"))
			})

			It("returns an error", func() {
				err := migrator.Migrate(dstStore, srcStore)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("list-error"))
			})
		})
	})
})
