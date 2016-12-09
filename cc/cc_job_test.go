package cc_test

import (
	. "github.com/c0-ops/goblob/cc"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/c0-ops/goblob/bosh/fakes"
)

var _ = Describe("CloudController Job", func() {

	var(
		cc *CloudController
		boshClient fakes.FakeBoshClient
	)
	BeforeEach(func() {
		boshClient = fakes.NewFakeBoshClient()
		cc = NewCloudController(boshClient, []CCJob{
			CCJob{
				Name: "cloud_controller",
				Deployment: "cf-deployment",
				Index: 0,
			},
			CCJob{
				Name: "cloud_controller",
				Deployment: "cf-deployment",
				Index: 1,
			},
		},)
	})

	Describe("Stop", func() {

		It("should stop the bosh cc job", func() {
			err := cc.Stop()
			Expect(err).Should(BeNil())
			Expect(cc.GetStatus()).Should(Equal("stopped"))
		})
	})

	Describe("Start", func() {

		It("should start the bosh cc job", func() {
			err := cc.Start()
			Expect(err).Should(BeNil())
			Expect(cc.GetStatus()).Should(Equal("started"))
		})
	})

})
