package nfs_test

import (
	"errors"

	. "github.com/c0-ops/goblob/nfs"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/c0-ops/goblob"
)

var _ = Describe("NFSStore", func() {
	var store goblob.Store
	BeforeEach(func() {
		store = New("fixtures")
	})

	Describe("List()", func() {
		It("Should return a list of blobs", func() {
			blobs, err := store.List()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(len(blobs)).Should(BeEquivalentTo(7))
			Ω(blobs[0].Filename).Should(BeEquivalentTo("ea07de9b-dd94-477c-b904-0f77d47dd111_a32d9ae40371d557c7c90eb2affc3d7bba6abe69"))
			Ω(blobs[0].Checksum).Should(BeEquivalentTo("b026324c6904b2a9cb4b88d6d61c81d1"))
			Ω(blobs[0].Path).Should(BeEquivalentTo("cc-buildpacks/ea/07"))
		})
	})
	Describe("Read()", func() {
		It("Given a file it should return a reader", func() {
			blobs, err := store.List()
			Ω(err).ShouldNot(HaveOccurred())

			reader, err := store.Read(blobs[0])
			Ω(err).ShouldNot(HaveOccurred())
			Ω(reader).ShouldNot(BeNil())
		})
	})
	Describe("Write()", func() {
		It("Should return an error", func() {
			err := errors.New("writing to the NFS store is not supported")
			Ω(store.Write(&goblob.Blob{}, nil)).Should(BeEquivalentTo(err))
		})
	})
})
