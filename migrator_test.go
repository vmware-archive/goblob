package goblob_test

import (
	"bytes"
	"errors"

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
		m = &CloudFoundryMigrator{}
		cf = &mock.FakeCloudFoundry{}
		dstStore = &mock.FakeStore{}
		srcStore = &mock.FakeStore{}
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
			srcStore.ListReturns([]Blob{Blob{
				Filename: "aabbfile",
				Checksum: "5d41402abc4b2a76b9719d911017c592",
				Path:     "/var/vcap/store/shared/cc-buildpacks/aa/bb",
			}}, nil)
			reader := bytes.NewReader([]byte("hello"))
			srcStore.ReadReturns(reader, nil)
			dstStore.WriteReturns(nil)
			dstStore.ReadReturns(reader, nil)

			err := m.Migrate(dstStore, srcStore)
			Ω(err).Should(BeNil())
			Ω(srcStore.ListCallCount()).Should(BeEquivalentTo(1))
			Ω(srcStore.ReadCallCount()).Should(BeEquivalentTo(1))
			Ω(dstStore.WriteCallCount()).Should(BeEquivalentTo(1))
			Ω(dstStore.ReadCallCount()).Should(BeEquivalentTo(1))
			writeBlob, writeReader := dstStore.WriteArgsForCall(0)
			Ω(writeBlob).ShouldNot(BeNil())
			Ω(writeReader).To(Equal(reader))
		})

		It("Should error on read from source", func() {
			controlErr := errors.New("got an error")
			cf.StoreReturns(srcStore, nil)
			srcStore.ListReturns([]Blob{Blob{
				Filename: "aabbfile",
				Checksum: "5d41402abc4b2a76b9719d911017c592",
				Path:     "/var/vcap/store/shared/cc-buildpacks/aa/bb",
			}}, nil)
			reader := bytes.NewReader([]byte("hello"))
			srcStore.ReadReturns(reader, controlErr)

			err := m.Migrate(dstStore, srcStore)
			Ω(err).Should(BeEquivalentTo(controlErr))
			Ω(srcStore.ListCallCount()).Should(BeEquivalentTo(1))
			Ω(srcStore.ReadCallCount()).Should(BeEquivalentTo(1))
		})

		It("Should error on write", func() {
			controlErr := errors.New("got an error")
			cf.StoreReturns(srcStore, nil)
			srcStore.ListReturns([]Blob{Blob{
				Filename: "aabbfile",
				Checksum: "5d41402abc4b2a76b9719d911017c592",
				Path:     "/var/vcap/store/shared/cc-buildpacks/aa/bb",
			}}, nil)
			reader := bytes.NewReader([]byte("hello"))
			srcStore.ReadReturns(reader, nil)
			dstStore.WriteReturns(controlErr)
			dstStore.ReadReturns(reader, nil)

			err := m.Migrate(dstStore, srcStore)
			Ω(err).Should(BeEquivalentTo(controlErr))
			Ω(srcStore.ListCallCount()).Should(BeEquivalentTo(1))
			Ω(srcStore.ReadCallCount()).Should(BeEquivalentTo(1))
			Ω(dstStore.WriteCallCount()).Should(BeEquivalentTo(1))
			writeBlob, writeReader := dstStore.WriteArgsForCall(0)
			Ω(writeBlob).ShouldNot(BeNil())
			Ω(writeReader).To(Equal(reader))
		})

		It("Should error on destination read", func() {
			controlErr := errors.New("got an error")
			cf.StoreReturns(srcStore, nil)
			srcStore.ListReturns([]Blob{Blob{
				Filename: "aabbfile",
				Checksum: "5d41402abc4b2a76b9719d911017c592",
				Path:     "/var/vcap/store/shared/cc-buildpacks/aa/bb",
			}}, nil)
			reader := bytes.NewReader([]byte("hello"))
			srcStore.ReadReturns(reader, nil)
			dstStore.WriteReturns(nil)
			dstStore.ReadReturns(reader, controlErr)

			err := m.Migrate(dstStore, srcStore)
			Ω(err).Should(BeEquivalentTo(controlErr))
			Ω(srcStore.ListCallCount()).Should(BeEquivalentTo(1))
			Ω(srcStore.ReadCallCount()).Should(BeEquivalentTo(1))
			Ω(dstStore.WriteCallCount()).Should(BeEquivalentTo(1))
			writeBlob, writeReader := dstStore.WriteArgsForCall(0)
			Ω(writeBlob).ShouldNot(BeNil())
			Ω(writeReader).To(Equal(reader))
		})

		It("Should error on checksum mismatch", func() {
			controlErr := errors.New("Checksum [5d41402abc4b2a76b9719d911017c592] does not match [abcd]")
			cf.StoreReturns(srcStore, nil)
			srcStore.ListReturns([]Blob{Blob{
				Filename: "aabbfile",
				Checksum: "abcd",
				Path:     "/var/vcap/store/shared/cc-buildpacks/aa/bb",
			}}, nil)
			reader := bytes.NewReader([]byte("hello"))
			srcStore.ReadReturns(reader, nil)
			dstStore.WriteReturns(nil)
			dstStore.ReadReturns(reader, nil)

			err := m.Migrate(dstStore, srcStore)
			Ω(err).Should(BeEquivalentTo(controlErr))
			Ω(srcStore.ListCallCount()).Should(BeEquivalentTo(1))
			Ω(srcStore.ReadCallCount()).Should(BeEquivalentTo(1))
			Ω(dstStore.WriteCallCount()).Should(BeEquivalentTo(1))
			Ω(dstStore.ReadCallCount()).Should(BeEquivalentTo(1))
			writeBlob, writeReader := dstStore.WriteArgsForCall(0)
			Ω(writeBlob).ShouldNot(BeNil())
			Ω(writeReader).To(Equal(reader))
		})
		// It("Should return an error when there is a checksum mismatch", func() {
		// 	cf.StoreReturns(srcStore, nil)
		// 	cf.IdentifierReturns("0987654321")
		// 	srcStore.ListReturns([]Blob{Blob{
		// 		Filename: "aabbfile",
		// 		Checksum: "1234567890",
		// 		Path:     "cc-buildpacks/aa/bb",
		// 	}}, nil)
		// 	reader := bytes.NewReader([]byte("hello"))
		// 	srcStore.ReadReturns(reader, nil)
		// 	dstStore.WriteReturns(nil)
		//
		// 	err := m.Migrate(dstStore, cf)
		// 	Ω(err).ShouldNot(BeNil())
		// 	Ω(srcStore.ListCallCount()).Should(BeEquivalentTo(1))
		// 	Ω(srcStore.ReadCallCount()).Should(BeEquivalentTo(1))
		// 	Ω(dstStore.WriteCallCount()).Should(BeEquivalentTo(1))
		// 	writeBlob, writeReader := dstStore.WriteArgsForCall(0)
		// 	Ω(writeBlob).ShouldNot(BeNil())
		// 	Ω(writeBlob.Filename).Should(BeEquivalentTo("aabbfile"))
		// 	Ω(writeBlob.Checksum).Should(BeEquivalentTo("1234567890"))
		// 	Ω(writeBlob.Path).Should(BeEquivalentTo("cc-buildpacks-0987654321/aa/bb"))
		// 	Ω(writeReader).To(Equal(reader))
		// })
	})
})
