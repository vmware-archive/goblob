package validation_test

import (
	"path"

	. "github.com/c0-ops/goblob/validation"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Md5", func() {
	It("Generates correct checksums", func() {
		checksum, err := Checksum(path.Join(".", "fixtures", "testfile"))
		Ω(err).Should(BeNil())
		Ω(checksum).Should(BeEquivalentTo("b026324c6904b2a9cb4b88d6d61c81d1"))
	})

	It("Returns an error for a missing filename", func() {
		checksum, err := Checksum(path.Join(".", "fixtures", "testmissing"))
		Ω(err).ShouldNot(BeNil())
		Ω(checksum).Should(BeEquivalentTo(""))
	})
})
