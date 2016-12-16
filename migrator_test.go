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
			It("Should return an error if the cloud foundry is empty", func() {
				var emptyStore = errors.New("cloud foundry is empty")
				err := m.Migrate(dstStore, nil)
				Ω(err).Should(Equal(emptyStore))
			})

			It("Should return an error if the cloud foundry store errors", func() {
				var theErr = errors.New("error retrieving store")
				cf.StoreReturns(nil, theErr)
				err := m.Migrate(dstStore, cf)
				Ω(err).Should(Equal(theErr))
			})

			It("Should return an error if the destination store is nil", func() {
				var emptyStore = errors.New("src is an empty store")
				err := m.Migrate(dstStore, cf)
				Ω(err).Should(Equal(emptyStore))
			})

			It("Should return an error if the destination store is nil", func() {
				var emptyStore = errors.New("dst is an empty store")
				cf.StoreReturns(srcStore, nil)
				err := m.Migrate(nil, cf)
				Ω(err).Should(Equal(emptyStore))
			})

			It("Should return an error if the source store has no files", func() {
				var emptyStore = errors.New("the source store has no files")
				cf.StoreReturns(srcStore, nil)
				srcStore.ListReturns(nil, nil)
				err := m.Migrate(dstStore, cf)
				Ω(err).Should(Equal(emptyStore))
				Ω(srcStore.ListCallCount()).Should(BeEquivalentTo(1))
			})

			It("Should error if the source store errors on List", func() {
				var testErr = errors.New("test")
				cf.StoreReturns(srcStore, nil)
				srcStore.ListReturns(nil, testErr)
				err := m.Migrate(dstStore, cf)
				Ω(err).Should(Equal(testErr))
				Ω(srcStore.ListCallCount()).Should(BeEquivalentTo(1))
			})
		})
	})

	Describe("When the source store has files", func() {
		It("Should retrieve a file if the source store has one", func() {
			cf.StoreReturns(srcStore, nil)
			srcStore.ListReturns([]Blob{Blob{
				Filename: "test",
				Checksum: "123456789",
				Path:     "root/src/file",
			}}, nil)
			srcStore.ReadReturns(nil, nil)

			err := m.Migrate(dstStore, cf)
			Ω(err).Should(BeNil())
			Ω(srcStore.ListCallCount()).Should(BeEquivalentTo(1))
			Ω(srcStore.ReadCallCount()).Should(BeEquivalentTo(1))
		})

		It("Should call destination write", func() {
			cf.StoreReturns(srcStore, nil)
			srcStore.ListReturns([]Blob{Blob{
				Filename: "aabbfile",
				Checksum: "123456789",
				Path:     "/var/vcap/store/shared/cc-buildpacks/aa/bb",
			}}, nil)
			reader := bytes.NewReader([]byte("hello"))
			srcStore.ReadReturns(reader, nil)
			dstStore.WriteReturns(nil)

			err := m.Migrate(dstStore, cf)
			Ω(err).Should(BeNil())
			Ω(srcStore.ListCallCount()).Should(BeEquivalentTo(1))
			Ω(srcStore.ReadCallCount()).Should(BeEquivalentTo(1))
			Ω(dstStore.WriteCallCount()).Should(BeEquivalentTo(1))
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
