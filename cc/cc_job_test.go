package cc_test

import (
	. "github.com/c0-ops/goblob/cc"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CloudController Job", func() {

	var(
		cc CloudController
	)
	BeforeEach(func() {
		cc = NewCloudController()
	})

	Describe("Stop", func() {

		It("should stop the bosh cc job", func() {
			cc.Stop()
			Expect(cc.Status).ShouldNot(Equal("started"))
		})
	})

	Describe("Start", func() {

		It("should start the bosh cc job", func() {
			cc.Start()
			Expect(cc.Status).ShouldNot(Equal("stopped"))
		})
	})

})
