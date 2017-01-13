package validation_test

import (
	"path"

	. "github.com/pivotalservices/goblob/validation"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Md5", func() {
	It("Generates correct checksums", func() {
		checksum, err := Checksum(path.Join(".", "fixtures", "testfile"))
		Ω(err).Should(BeNil())
		Ω(checksum).Should(BeEquivalentTo("b026324c6904b2a9cb4b88d6d61c81d1"))
	})

	It("Generates correct checksums", func() {
		checksum, err := Checksum(path.Join(".", "fixtures", "013110a30e2a475551c801b4c45e497ce71c26fe"))
		Ω(err).Should(BeNil())
		Ω(checksum).Should(BeEquivalentTo("9e63a667623321944e174d3d3ea16e9e"))
	})

	It("Returns an error for a missing filename", func() {
		checksum, err := Checksum(path.Join(".", "fixtures", "testmissing"))
		Ω(err).ShouldNot(BeNil())
		Ω(checksum).Should(BeEquivalentTo(""))
	})
})
