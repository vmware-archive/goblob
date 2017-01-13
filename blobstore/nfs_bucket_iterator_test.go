package blobstore_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pivotalservices/goblob/blobstore"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("NFSBucketIterator", func() {
	var (
		iterator blobstore.BucketIterator
		store    blobstore.Blobstore
		baseDir  string
	)

	BeforeEach(func() {
		var err error
		baseDir, err = ioutil.TempDir("", "nfs-bucket-iterator-test")
		Expect(err).NotTo(HaveOccurred())

		store = blobstore.NewNFS(baseDir)

		err = os.MkdirAll(filepath.Join(baseDir, "some-bucket"), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		iterator, err = store.NewBucketIterator("some-bucket")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		os.RemoveAll(baseDir)
	})

	Describe("Next", func() {
		It("returns an error", func() {
			_, err := iterator.Next()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("no more items in iterator"))
		})

		Context("when a blob exists in the bucket", func() {
			var expectedBlob blobstore.Blob

			BeforeEach(func() {
				expectedBlob = blobstore.Blob{
					Path: "some-bucket/some-path/some-file",
				}

				err := os.MkdirAll(filepath.Join(baseDir, "some-bucket", "some-path"), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				err = ioutil.WriteFile(
					filepath.Join(baseDir, "some-bucket", "some-path", "some-file"),
					[]byte("content"),
					os.ModePerm,
				)
				Expect(err).NotTo(HaveOccurred())

				iterator, err = store.NewBucketIterator("some-bucket")
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns the blob", func() {
				blob, err := iterator.Next()
				Expect(err).NotTo(HaveOccurred())
				Expect(*blob).To(Equal(expectedBlob))
			})

			It("returns an error when all blobs have been listed", func() {
				_, err := iterator.Next()
				Expect(err).NotTo(HaveOccurred())

				_, err = iterator.Next()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("no more items in iterator"))
			})
		})

		Context("when a file named .nfs_test exists in the bucket", func() {
			BeforeEach(func() {
				err := os.MkdirAll(filepath.Join(baseDir, "some-bucket", "some-path"), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				err = ioutil.WriteFile(
					filepath.Join(baseDir, "some-bucket", "some-path", ".nfs_test"),
					[]byte{},
					os.ModePerm,
				)
				Expect(err).NotTo(HaveOccurred())

				err = ioutil.WriteFile(
					filepath.Join(baseDir, "some-bucket", "some-path", "some-file"),
					[]byte("content"),
					os.ModePerm,
				)
				Expect(err).NotTo(HaveOccurred())

				iterator, err = store.NewBucketIterator("some-bucket")
				Expect(err).NotTo(HaveOccurred())
			})

			It("skips the file", func() {
				_, err := iterator.Next()
				Expect(err).NotTo(HaveOccurred())

				_, err = iterator.Next()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("no more items in iterator"))
			})
		})
	})

	Describe("Done", func() {
		Context("when blobs exist in the bucket", func() {
			BeforeEach(func() {
				err := os.MkdirAll(filepath.Join(baseDir, "some-bucket", "some-path"), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				err = ioutil.WriteFile(
					filepath.Join(baseDir, "some-bucket", "some-path", "some-file"),
					[]byte("content"),
					os.ModePerm,
				)
				Expect(err).NotTo(HaveOccurred())

				err = ioutil.WriteFile(
					filepath.Join(baseDir, "some-bucket", "some-path", "some-other-file"),
					[]byte("content"),
					os.ModePerm,
				)
				Expect(err).NotTo(HaveOccurred())

				iterator, err = store.NewBucketIterator("some-bucket")
				Expect(err).NotTo(HaveOccurred())
			})

			It("causes Next to return an error", func() {
				iterator.Done()

				_, err := iterator.Next()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("no more items in iterator"))
			})
		})
	})
})
