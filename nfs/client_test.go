package nfs_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/c0-ops/goblob/nfs"
	"github.com/c0-ops/goblob/nfs/fakes"
	"github.com/c0-ops/goblob/cmd"
	"code.cloudfoundry.org/lager"
)

var _ = Describe("nfs client", func() {
	Describe("NewNFSClient", func() {
		var logger lager.Logger
		BeforeEach(func() {
			logger = lager.NewLogger("logger")
		})

		Context("when executer is created successfully", func() {
			var origExecuterFunction func(cmd.SshConfig) (cmd.Executor, error)
			var logger lager.Logger

			BeforeEach(func() {
				logger = lager.NewLogger("logger")
				origExecuterFunction = SshCmdExecutor
				SshCmdExecutor = func(cmd.SshConfig) (cmd.Executor, error) {
					return &fakes.SuccessMockNFSExecuter{}, nil
				}
			})

			AfterEach(func() {
				SshCmdExecutor = origExecuterFunction
			})

			It("should return a nil error and a valid nfs client", func() {
				n, err := NewNFSClient("vcap", "pass", "0.0.0.0", "/var/somepath", nil, logger)
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
				n, err := NewNFSClient("vcap", "pass", "0.0.0.0", "/var/somepath", nil, logger)
				Expect(err).ShouldNot(BeNil())
				Expect(n).Should(BeNil())
			})
		})
	})
})
