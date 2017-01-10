package goblob_test

import (
	"bytes"
	"errors"
	"io/ioutil"

	. "github.com/c0-ops/goblob"
	"github.com/c0-ops/goblob/mock"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Migrator", func() {
	var m *CloudFoundryMigrator
	var cf *mock.FakeCloudFoundry
	var dstStore *mock.FakeStore
	var srcStore *mock.FakeStore

	BeforeEach(func() {
		m = New(20)
		cf = &mock.FakeCloudFoundry{}
		dstStore = &mock.FakeStore{}
		srcStore = &mock.FakeStore{}
	})

	Describe("BlobMigrator", func() {
		var controlBlob *Blob
		var err error

		Context("when calling SingleBlobError", func() {
			controlFilename := "filename"
			controlFilepath := "pathapth"
			controlErrorMessage := "messsage of something bad"
			It("should compose a new error message from blob info and error info given", func() {
				bm := new(BlobMigrate)
				err := bm.SingleBlobError(&Blob{
					Filename: controlFilename,
					Path:     controlFilepath,
				}, errors.New(controlErrorMessage))
				Ω(err.Error()).Should(ContainSubstring(controlFilename))
				Ω(err.Error()).Should(ContainSubstring(controlFilepath))
				Ω(err.Error()).Should(ContainSubstring(controlErrorMessage))
			})
		})

		Context("when called on a valid src, dst and blob set", func() {
			BeforeEach(func() {
				controlBlob = &Blob{
					Filename: "aabbfile",
					Checksum: "5d41402abc4b2a76b9719d911017c592",
					Path:     "/var/vcap/store/shared/cc-buildpacks/aa/bb",
				}
				cf.StoreReturns(srcStore, nil)
				srcStore.ListReturns([]*Blob{controlBlob}, nil)
				reader := ioutil.NopCloser(bytes.NewReader([]byte("hello")))
				srcStore.ReadReturns(reader, nil)
				dstStore.WriteReturns(nil)
				dstStore.ChecksumReturns(controlBlob.Checksum, nil)
				bm := new(BlobMigrate)
				bm.Init(dstStore, srcStore)
				err = bm.MigrateSingleBlob(controlBlob)
			})
			It("Should complete successfully", func() {
				Ω(err).ShouldNot(HaveOccurred())
			})
		})

		Context("when called on a src/dst/blob set which can not be migrated", func() {
			BeforeEach(func() {
				controlBlob = &Blob{
					Filename: "aabbfile",
					Checksum: "5d41402abc4b2a76b9719d911017c592",
					Path:     "/var/vcap/store/shared/cc-buildpacks/aa/bb",
				}
				cf.StoreReturns(srcStore, nil)
				reader := ioutil.NopCloser(bytes.NewReader([]byte("hello")))
				srcStore.ReadReturns(reader, nil)
			})

			Context("when we can not read from the source", func() {
				var controlMessage = "something is wrong with reading"
				var controlError = errors.New(controlMessage)
				BeforeEach(func() {
					srcStore.ReadReturns(nil, controlError)
					bm := new(BlobMigrate)
					bm.Init(dstStore, srcStore)
					err = bm.MigrateSingleBlob(controlBlob)
				})

				It("should yield an error", func() {
					Ω(err).Should(HaveOccurred())
					Ω(err.Error()).Should(ContainSubstring(controlMessage))
				})

				It("should add the filename to the error message", func() {
					Ω(err.Error()).Should(ContainSubstring(controlBlob.Filename))
					Ω(err.Error()).Should(ContainSubstring(controlBlob.Path))
				})
			})

			Context("when we can not write to the destination", func() {
				var controlMessage = "something is wrong with writing"
				var controlError = errors.New(controlMessage)
				BeforeEach(func() {
					dstStore.WriteReturns(controlError)
					bm := new(BlobMigrate)
					bm.Init(dstStore, srcStore)
					err = bm.MigrateSingleBlob(controlBlob)
				})
				It("should yield an error", func() {
					Ω(err).Should(HaveOccurred())
					Ω(err.Error()).Should(ContainSubstring(controlMessage))
				})

				It("should add the filename to the error message", func() {
					Ω(err.Error()).Should(ContainSubstring(controlBlob.Filename))
					Ω(err.Error()).Should(ContainSubstring(controlBlob.Path))
				})
			})

			Context("when we can not read the checksum", func() {
				var controlMessage = "something is wrong with reading the checksum"
				var controlError = errors.New(controlMessage)
				BeforeEach(func() {
					dstStore.ChecksumReturns("", controlError)
					bm := new(BlobMigrate)
					bm.Init(dstStore, srcStore)
					err = bm.MigrateSingleBlob(controlBlob)
				})
				It("should yield an error", func() {
					Ω(err).Should(HaveOccurred())
					Ω(err.Error()).Should(ContainSubstring(controlMessage))
				})

				It("should add the filename to the error message", func() {
					Ω(err.Error()).Should(ContainSubstring(controlBlob.Filename))
					Ω(err.Error()).Should(ContainSubstring(controlBlob.Path))
				})
			})

			Context("when the checksums do not match", func() {
				BeforeEach(func() {
					srcStore.ListReturns([]*Blob{controlBlob}, nil)
					dstStore.ListReturns([]*Blob{controlBlob}, nil)
					reader := ioutil.NopCloser(bytes.NewReader([]byte("hello")))
					srcStore.ReadReturns(reader, nil)
					dstStore.WriteReturns(nil)
					dstStore.ReadReturns(reader, nil)
					bm := new(BlobMigrate)
					bm.Init(dstStore, srcStore)
					err = bm.MigrateSingleBlob(controlBlob)
				})
				It("should yield an error", func() {
					Ω(err).Should(HaveOccurred())
					Ω(err.Error()).Should(ContainSubstring("Checksum [] does not match"))
				})

				It("should add the filename to the error message", func() {
					Ω(err.Error()).Should(ContainSubstring(controlBlob.Filename))
					Ω(err.Error()).Should(ContainSubstring(controlBlob.Path))
				})
			})
		})
	})

	Describe("When the source store has no files", func() {
		Describe("Migrate(store, cf)", func() {

			It("Should return an error if the destination store is nil", func() {
				var emptyStore = errors.New("src is an empty store")
				err := m.Migrate(dstStore, nil)
				Ω(err).Should(Equal(emptyStore))
			})

			It("Should return an error if the destination store is nil", func() {
				var emptyStore = errors.New("dst is an empty store")
				err := m.Migrate(nil, srcStore)
				Ω(err).Should(Equal(emptyStore))
			})

			It("Should return an error if the source store has no files", func() {
				var emptyStore = errors.New("the source store has no files")
				cf.StoreReturns(srcStore, nil)
				srcStore.ListReturns(nil, nil)
				err := m.Migrate(dstStore, srcStore)
				Ω(err).Should(Equal(emptyStore))
				Ω(srcStore.ListCallCount()).Should(BeEquivalentTo(1))
			})

			It("Should error if the source store errors on List", func() {
				var testErr = errors.New("test")
				cf.StoreReturns(srcStore, nil)
				srcStore.ListReturns(nil, testErr)
				err := m.Migrate(dstStore, srcStore)
				Ω(err).Should(Equal(testErr))
				Ω(srcStore.ListCallCount()).Should(BeEquivalentTo(1))
			})
		})
	})

	Describe("When the source store has files", func() {
		It("Should successfully migrate", func() {
			cf.StoreReturns(srcStore, nil)
			srcStore.ListReturns([]*Blob{&Blob{
				Filename: "aabbfile",
				Checksum: "5d41402abc4b2a76b9719d911017c592",
				Path:     "/var/vcap/store/shared/cc-buildpacks/aa/bb",
			}}, nil)
			reader := ioutil.NopCloser(bytes.NewReader([]byte("hello")))
			srcStore.ReadReturns(reader, nil)
			dstStore.WriteReturns(nil)
			dstStore.ChecksumReturns("5d41402abc4b2a76b9719d911017c592", nil)

			err := m.Migrate(dstStore, srcStore)
			Ω(err).Should(BeNil())
			Ω(srcStore.ListCallCount()).Should(BeEquivalentTo(1))
			Ω(srcStore.ReadCallCount()).Should(BeEquivalentTo(1))
			Ω(dstStore.WriteCallCount()).Should(BeEquivalentTo(1))
			Ω(dstStore.ChecksumCallCount()).Should(BeEquivalentTo(1))
			writeBlob, writeReader := dstStore.WriteArgsForCall(0)
			Ω(writeBlob).ShouldNot(BeNil())
			Ω(writeReader).To(Equal(reader))
		})

		It("Should error on read from source", func() {
			controlErrorMessage := "got an error"
			controlErr := errors.New(controlErrorMessage)
			cf.StoreReturns(srcStore, nil)
			srcStore.ListReturns([]*Blob{&Blob{
				Filename: "aabbfile",
				Checksum: "5d41402abc4b2a76b9719d911017c592",
				Path:     "/var/vcap/store/shared/cc-buildpacks/aa/bb",
			}}, nil)
			reader := ioutil.NopCloser(bytes.NewReader([]byte("hello")))
			srcStore.ReadReturns(reader, controlErr)

			err := m.Migrate(dstStore, srcStore)
			Ω(err.Error()).Should(ContainSubstring(controlErrorMessage))
			Ω(srcStore.ListCallCount()).Should(BeEquivalentTo(1))
			Ω(srcStore.ReadCallCount()).Should(BeEquivalentTo(1))
		})

		It("Should error on write", func() {
			controlErrorMessage := "got an error"
			controlErr := errors.New(controlErrorMessage)
			cf.StoreReturns(srcStore, nil)
			srcStore.ListReturns([]*Blob{&Blob{
				Filename: "aabbfile",
				Checksum: "5d41402abc4b2a76b9719d911017c592",
				Path:     "/var/vcap/store/shared/cc-buildpacks/aa/bb",
			}}, nil)
			reader := ioutil.NopCloser(bytes.NewReader([]byte("hello")))
			srcStore.ReadReturns(reader, nil)
			dstStore.WriteReturns(controlErr)
			dstStore.ReadReturns(reader, nil)

			err := m.Migrate(dstStore, srcStore)
			Ω(err.Error()).Should(ContainSubstring(controlErrorMessage))
			Ω(srcStore.ListCallCount()).Should(BeEquivalentTo(1))
			Ω(srcStore.ReadCallCount()).Should(BeEquivalentTo(1))
			Ω(dstStore.WriteCallCount()).Should(BeEquivalentTo(1))
			writeBlob, writeReader := dstStore.WriteArgsForCall(0)
			Ω(writeBlob).ShouldNot(BeNil())
			Ω(writeReader).To(Equal(reader))
		})

		It("Should error on checksum mismatch", func() {
			controlErr := errors.New("error at /var/vcap/store/shared/cc-buildpacks/aa/bb/aabbfile: Checksum [5d41402abc4b2a76b9719d911017c592] does not match [abcd]")
			cf.StoreReturns(srcStore, nil)
			srcStore.ListReturns([]*Blob{&Blob{
				Filename: "aabbfile",
				Checksum: "abcd",
				Path:     "/var/vcap/store/shared/cc-buildpacks/aa/bb",
			}}, nil)
			reader := ioutil.NopCloser(bytes.NewReader([]byte("hello")))
			srcStore.ReadReturns(reader, nil)
			dstStore.WriteReturns(nil)
			dstStore.ChecksumReturns("5d41402abc4b2a76b9719d911017c592", nil)

			err := m.Migrate(dstStore, srcStore)
			Ω(err).Should(BeEquivalentTo(controlErr))
			Ω(srcStore.ListCallCount()).Should(BeEquivalentTo(1))
			Ω(srcStore.ReadCallCount()).Should(BeEquivalentTo(1))
			Ω(dstStore.WriteCallCount()).Should(BeEquivalentTo(1))
			Ω(dstStore.ChecksumCallCount()).Should(BeEquivalentTo(1))
			writeBlob, writeReader := dstStore.WriteArgsForCall(0)
			Ω(writeBlob).ShouldNot(BeNil())
			Ω(writeReader).To(Equal(reader))
		})

		It("not migrate already migrated files", func() {
			cf.StoreReturns(srcStore, nil)
			srcStore.ListReturns([]*Blob{&Blob{
				Filename: "aabbfile",
				Checksum: "5d41402abc4b2a76b9719d911017c592",
				Path:     "/var/vcap/store/shared/cc-buildpacks/aa/bb",
			}}, nil)

			dstStore.ExistsStub = func(blob *Blob) bool {
				return blob.Filename == "aabbfile"
			}

			reader := ioutil.NopCloser(bytes.NewReader([]byte("hello")))
			srcStore.ReadReturns(reader, nil)
			dstStore.WriteReturns(nil)
			dstStore.ReadReturns(reader, nil)

			err := m.Migrate(dstStore, srcStore)
			Ω(err).Should(BeNil())
			Ω(srcStore.ListCallCount()).Should(BeEquivalentTo(1))
			Ω(srcStore.ReadCallCount()).Should(BeEquivalentTo(0))
		})

	})
})
