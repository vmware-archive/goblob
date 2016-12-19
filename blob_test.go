package goblob_test

import (
	. "github.com/c0-ops/goblob"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Blob", func() {

	Describe("When comparing blobs", func() {
		Describe("Equal", func() {
			It("Should return true", func() {
				srcBlob := Blob{
					Filename: "a",
					Checksum: "123",
				}
				targetBlob := Blob{
					Filename: "a",
					Checksum: "123",
				}
				Ω(srcBlob.Equal(targetBlob)).Should(BeTrue())
			})
			It("Should return false with different checksums", func() {
				srcBlob := Blob{
					Filename: "a",
					Checksum: "1234",
				}
				targetBlob := Blob{
					Filename: "a",
					Checksum: "123",
				}
				Ω(srcBlob.Equal(targetBlob)).Should(BeFalse())
			})
			It("Should return false with different filenames", func() {
				srcBlob := Blob{
					Filename: "b",
					Checksum: "123",
				}
				targetBlob := Blob{
					Filename: "a",
					Checksum: "123",
				}
				Ω(srcBlob.Equal(targetBlob)).Should(BeFalse())
			})

		})
	})

})
