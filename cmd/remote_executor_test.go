package cmd_test

import (
	"bytes"
	"errors"
	"io"
	"strings"

	"golang.org/x/crypto/ssh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-customer0/cfblobmigrator/cmd"
)

type mockClient struct {
	session SSHSession
}

func (c *mockClient) NewSession() (SSHSession, error) {
	return c.session, nil
}

type mockSession struct {
	StartSuccess  bool
	StdOutSuccess bool
	WaitSuccess   bool
	CloseSuccess  bool
}

func (session *mockSession) Start(command string) (err error) {
	if !session.StartSuccess {
		err = errors.New("")
	}
	return
}

func (session *mockSession) Close() (err error) {
	if !session.CloseSuccess {
		err = errors.New("")
	}
	return
}

func (session *mockSession) Wait() (err error) {
	if !session.WaitSuccess {
		err = errors.New("")
	}
	return
}

func (session *mockSession) StdoutPipe() (reader io.Reader, err error) {
	if !session.StdOutSuccess {
		err = errors.New("")
		return nil, err
	}
	reader = strings.NewReader("mocksession")
	return
}

var _ = Describe("Ssh", func() {
	var (
		session *mockSession
		client  *mockClient
	)

	BeforeEach(func() {
		session = &mockSession{StartSuccess: true,
			StdOutSuccess: true,
			WaitSuccess:   true,
			CloseSuccess:  true}
		client = &mockClient{session: session}

	})
	Describe("NewRemoteExecutor", func() {
		Describe("given it creates a *DefaultRemoteExecutor to satisfy the Executer interface", func() {
			Context("when initializing the *DefaultRemoteExecutor", func() {
				It("then it should not dial the ssh connection, but lazy load it", func() {
					executor, err := NewRemoteExecutor(SshConfig{})
					Ω(err).ShouldNot(HaveOccurred())
					Ω(executor.(*DefaultRemoteExecutor).Client).Should(BeNil())
				})
			})
		})
	})
	Describe("Given a faultRemoteExecutor", func() {
		Context("when calling Execute and a client connection can not be made", func() {
			var fakeExecutor *DefaultRemoteExecutor
			BeforeEach(func() {
				fakeExecutor = new(DefaultRemoteExecutor)
				fakeExecutor.LazyClientDial = func() {}
			})
			It("then we should error and exit gracefully", func() {
				Ω(func() {
					fakeExecutor.ExecuteForWrite(bytes.NewBufferString(""), "command string")
				}).ShouldNot(Panic())
			})
		})
	})

	Describe("SshConfig", func() {
		Describe("given a GetAuthMethod method", func() {
			keySigner, _ := ssh.ParsePrivateKey([]byte(``))
			var (
				config               *SshConfig
				controlAuthMethodKey = []ssh.AuthMethod{
					ssh.PublicKeys(keySigner),
				}
				controlAuthMethodPassword = []ssh.AuthMethod{
					ssh.Password(""),
				}
			)

			Context("when called on a sshconfig object containing a sslkey", func() {
				BeforeEach(func() {
					config = &SshConfig{
						SSLKey:   "random key data",
						Username: "randomuser",
						Host:     "asldkjfasd",
						Password: "xxxxxxx",
						Port:     8888,
					}
				})

				It("then it should return a authmethod which uses keypairs", func() {
					authMethod := config.GetAuthMethod()
					Ω(authMethod[0]).Should(BeAssignableToTypeOf(controlAuthMethodKey[0]))
					Ω(authMethod[0]).ShouldNot(BeAssignableToTypeOf(controlAuthMethodPassword[0]))
				})
			})
			Context("when called on a sshconfig object containing only a username/pass", func() {
				BeforeEach(func() {
					config = &SshConfig{
						Username: "randomuser",
						Host:     "asldkjfasd",
						Port:     8888,
						Password: "reallysecurestuff",
					}
				})
				It("then it should return a authmethod which uses username/pass auth", func() {
					authMethod := config.GetAuthMethod()
					Ω(authMethod[0]).Should(BeAssignableToTypeOf(controlAuthMethodPassword[0]))
					Ω(authMethod[0]).ShouldNot(BeAssignableToTypeOf(controlAuthMethodKey[0]))
				})
			})
		})
	})

	Describe("Session Run success", func() {
		Context("Everything is fine", func() {
			It("should write to the writer from the command output", func() {
				var writer bytes.Buffer
				executor := &DefaultRemoteExecutor{
					Client:         client,
					LazyClientDial: func() {},
				}
				executor.ExecuteForWrite(&writer, "command")
				Ω(writer.String()).Should(Equal("mocksession"))
			})
			It("should not return an error", func() {
				var writer bytes.Buffer
				executor := &DefaultRemoteExecutor{
					Client:         client,
					LazyClientDial: func() {},
				}
				err := executor.ExecuteForWrite(&writer, "command")
				Ω(err).ShouldNot(HaveOccurred())
			})
		})

	})
	Describe("Session Run failed", func() {

		Context("With bad stdpipeline", func() {
			It("should return an error on bad stdpipline", func() {
				var writer bytes.Buffer
				executor := &DefaultRemoteExecutor{
					Client:         client,
					LazyClientDial: func() {},
				}
				session.StdOutSuccess = false
				err := executor.ExecuteForWrite(&writer, "command")
				session.StdOutSuccess = false
				Ω(err).Should(HaveOccurred())
			})
		})
		Context("With bad command start", func() {
			It("should return an error", func() {
				var writer bytes.Buffer
				session.StartSuccess = false
				executor := &DefaultRemoteExecutor{
					Client:         client,
					LazyClientDial: func() {},
				}
				err := executor.ExecuteForWrite(&writer, "command")
				Ω(err).Should(HaveOccurred())
			})
		})
	})

})
