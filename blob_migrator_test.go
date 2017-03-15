package goblob_test

import (
	"errors"
	"io"
	"io/ioutil"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotalservices/goblob"
	"github.com/pivotalservices/goblob/blobstore"
	"github.com/pivotalservices/goblob/blobstore/blobstorefakes"
)

var _ = Describe("BlobMigrator", func() {
	var (
		blobMigrator goblob.BlobMigrator
		dstStore     *blobstorefakes.FakeBlobstore
		srcStore     *blobstorefakes.FakeBlobstore
	)

	BeforeEach(func() {
		dstStore = &blobstorefakes.FakeBlobstore{}
		srcStore = &blobstorefakes.FakeBlobstore{}
		blobMigrator = goblob.NewBlobMigrator(dstStore, srcStore)
	})

	Describe("Migrate", func() {
		var (
			controlBlob    *blobstore.Blob
			expectedReader io.ReadCloser
		)

		BeforeEach(func() {
			expectedReader = ioutil.NopCloser(strings.NewReader("some content"))
			srcStore.ReadReturns(expectedReader, nil)
			dstStore.ChecksumReturns("some-checksum", nil)
			controlBlob = &blobstore.Blob{
				Checksum: "some-checksum",
				Path:     "some-path/some-filename",
			}
		})

		It("tries to read the source blob", func() {
			err := blobMigrator.Migrate(controlBlob)
			Expect(err).NotTo(HaveOccurred())
			Expect(srcStore.ReadCallCount()).To(Equal(1))

			Expect(srcStore.ReadArgsForCall(0)).To(Equal(controlBlob))
		})

		It("tries to write the destination blob", func() {
			err := blobMigrator.Migrate(controlBlob)
			Expect(err).NotTo(HaveOccurred())

			Expect(dstStore.WriteCallCount()).To(Equal(1))

			blob, r := dstStore.WriteArgsForCall(0)
			Expect(blob).To(Equal(controlBlob))
			Expect(r).To(Equal(expectedReader))
		})

		It("tries to checksum the destination blob", func() {
			err := blobMigrator.Migrate(controlBlob)
			Expect(err).NotTo(HaveOccurred())
			Expect(dstStore.ChecksumCallCount()).To(Equal(1))

			Expect(dstStore.ChecksumArgsForCall(0)).To(Equal(controlBlob))
		})

		It("returns nil", func() {
			err := blobMigrator.Migrate(controlBlob)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when there is an error reading the source blob", func() {
			BeforeEach(func() {
				srcStore.ReadReturns(nil, errors.New("read-error"))
			})

			It("returns an error", func() {
				err := blobMigrator.Migrate(controlBlob)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("error reading blob at some-path/some-filename: read-error"))
			})
		})

		Context("when there is an error writing the destination blob", func() {
			BeforeEach(func() {
				dstStore.WriteReturns(errors.New("write-error"))
			})

			It("returns an error", func() {
				err := blobMigrator.Migrate(controlBlob)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("error writing blob at some-path/some-filename: write-error"))
			})
		})

		Context("when there is an error getting the destination checksum", func() {
			BeforeEach(func() {
				dstStore.ChecksumReturns("", errors.New("checksum-error"))
			})

			It("returns an error", func() {
				err := blobMigrator.Migrate(controlBlob)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("error checksumming blob at some-path/some-filename: checksum-error"))
			})
		})

		Context("when the checksums do not match", func() {
			BeforeEach(func() {
				dstStore.ChecksumReturns("other-checksum", nil)
			})

			It("returns an error", func() {
				err := blobMigrator.Migrate(controlBlob)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("error at some-path/some-filename: checksum [other-checksum] does not match [some-checksum]"))
			})
		})
	})
})
