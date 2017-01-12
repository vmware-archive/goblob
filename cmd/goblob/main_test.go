package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

func TestMain(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Main Suite")
}

var _ = Describe("Main", func() {
	AfterEach(func() {
		gexec.CleanupBuildArtifacts()
	})

	It("builds", func() {
		_, err := gexec.Build("github.com/c0-ops/goblob/cmd/goblob")
		Expect(err).NotTo(HaveOccurred())
	})
})
