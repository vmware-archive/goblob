package nfs_test

import (
	"fmt"
	"github.com/c0-ops/goblob/cmd"
	"github.com/c0-ops/goblob/cmd/fakes"
	. "github.com/c0-ops/goblob/nfs"
	faketar "github.com/c0-ops/goblob/tar/fakes"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("nfs client", func() {
	Describe("NewNFSClient", func() {
		var (
			logger    boshlog.Logger
			fs        *fakesys.FakeFileSystem
			extractor faketar.FakeCmdExtractor
		)
		BeforeEach(func() {
			fs = fakesys.NewFakeFileSystem()
			logger = boshlog.NewLogger(boshlog.LevelNone)
			extractor = faketar.NewFakeCmdExtractor()
		})

		Context("when executer is created successfully", func() {
			var origExecuterFunction func(cmd.SshConfig) (cmd.Executor, error)

			BeforeEach(func() {
				origExecuterFunction = SshCmdExecutor
				SshCmdExecutor = func(cmd.SshConfig) (cmd.Executor, error) {
					return &fakes.SuccessMockNFSExecuter{}, nil
				}
			})

			AfterEach(func() {
				SshCmdExecutor = origExecuterFunction
			})

			It("should return a nil error and a valid nfs client", func() {
				n, err := NewNFSClient("vcap", "pass", "0.0.0.0", extractor, fs, logger)
				Expect(err).Should(BeNil())
				Expect(n).ShouldNot(BeNil())
			})
		})

		Context("when executer fails to be created properly", func() {
			var origExecuterFunction func(cmd.SshConfig) (cmd.Executor, error)

			BeforeEach(func() {
				origExecuterFunction = SshCmdExecutor
				SshCmdExecutor = func(cmd.SshConfig) (ce cmd.Executor, err error) {
					ce = &fakes.FailureMockNFSExecuter{}
					err = fmt.Errorf("we have an error")
					return
				}
			})

			AfterEach(func() {
				SshCmdExecutor = origExecuterFunction
			})

			It("should return a nil error and nfs client", func() {
				n, err := NewNFSClient("vcap", "pass", "0.0.0.0", extractor, fs, logger)
				Expect(err).ShouldNot(BeNil())
				Expect(n).Should(BeNil())
			})
		})
	})
})
