package goblob_test

import (
	"errors"

	"code.cloudfoundry.org/workpool"

	"github.com/c0-ops/goblob"
	"github.com/c0-ops/goblob/blobstore"
	"github.com/c0-ops/goblob/blobstore/blobstorefakes"
	"github.com/c0-ops/goblob/goblobfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BlobstoreMigrator", func() {
	var (
		migrator     goblob.BlobstoreMigrator
		blobMigrator *goblobfakes.FakeBlobMigrator
		dstStore     *blobstorefakes.FakeBlobstore
		srcStore     *blobstorefakes.FakeBlobstore
		iterator     *blobstorefakes.FakeBucketIterator
	)

	BeforeEach(func() {
		dstStore = &blobstorefakes.FakeBlobstore{}
		srcStore = &blobstorefakes.FakeBlobstore{}
		blobMigrator = &goblobfakes.FakeBlobMigrator{}

		pool, err := workpool.NewWorkPool(1)
		Expect(err).NotTo(HaveOccurred())

		migrator = goblob.NewBlobstoreMigrator(pool, blobMigrator)

		iterator = &blobstorefakes.FakeBucketIterator{}
		srcStore.NewBucketIteratorReturns(iterator, nil)
	})

	Describe("Migrate", func() {
		var firstBlob, secondBlob, thirdBlob *blobstore.Blob

		BeforeEach(func() {
			firstBlob = &blobstore.Blob{
				Checksum: "some-file-checksum",
				Path:     "some-file-path/some-file",
			}

			secondBlob = &blobstore.Blob{
				Checksum: "some-other-file-checksum",
				Path:     "some-other-path/some-other-file",
			}

			thirdBlob = &blobstore.Blob{
				Checksum: "yet-another-file-checksum",
				Path:     "yet-another-path/yet-another-file",
			}

			iterator.CountReturns(3)
			iterator.NextStub = func() (*blobstore.Blob, error) {
				switch iterator.NextCallCount() {
				case 1:
					return firstBlob, nil
				case 2:
					return secondBlob, nil
				case 3:
					return thirdBlob, nil
				default:
					return nil, blobstore.ErrIteratorDone
				}
			}
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
				dstStore.ExistsStub = func(blob *blobstore.Blob) bool {
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
				blobMigrator.MigrateStub = func(blob *blobstore.Blob) error {
					if blob.Path == "some-other-path/some-other-file" {
						return errors.New("migrate-err")
					}
					return nil
				}
			})

			It("continues uploading", func() {
				err := migrator.Migrate(dstStore, srcStore)
				Expect(err).NotTo(HaveOccurred())

				Expect(blobMigrator.MigrateCallCount()).To(Equal(3))
				Expect(blobMigrator.MigrateArgsForCall(0)).To(Equal(firstBlob))
				Expect(blobMigrator.MigrateArgsForCall(1)).To(Equal(secondBlob))
				Expect(blobMigrator.MigrateArgsForCall(2)).To(Equal(thirdBlob))
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
				iterator.NextStub = func() (*blobstore.Blob, error) {
					return nil, errors.New("no more files!") // <-- this needs to be an exported err
				}
			})

			It("returns an error", func() {
				err := migrator.Migrate(dstStore, srcStore)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("no more files!"))
			})
		})

		XContext("when there is an error listing the source's files", func() {
			BeforeEach(func() {
				// srcStore.ListReturns(nil, errors.New("list-error"))
			})

			It("returns an error", func() {
				err := migrator.Migrate(dstStore, srcStore)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("list-error"))
			})
		})
	})
})
