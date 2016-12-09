package cmd

import (
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/pkg/sftp"
	"github.com/xchapter7x/lo"
)

//Taken from cfops

var DefaultPacketSize int = 1<<15

//SshConfig - for the SSH connection
type SshConfig struct {
	Username string
	Password string
	Host     string
	Port     int
	SSLKey   string
}

//GetAuthMethod -
func (s *SshConfig) GetAuthMethod() (authMethod []ssh.AuthMethod) {

	if s.SSLKey == "" {
		lo.G.Debug("using password for authn")
		authMethod = []ssh.AuthMethod{
			ssh.Password(s.Password),
		}

	} else {
		lo.G.Debug("using sslkey for authn")
		keySigner, _ := ssh.ParsePrivateKey([]byte(s.SSLKey))

		authMethod = []ssh.AuthMethod{
			ssh.PublicKeys(keySigner),
		}
	}
	return
}

//ClientInterface -
type ClientInterface interface {
	NewSession() (SSHSession, error)
}

//DefaultRemoteExecutor -
type DefaultRemoteExecutor struct {
	Client         ClientInterface
	LazyClientDial func()
	once           sync.Once
	sshclient *ssh.Client
}

//SshClientWrapper - of ssh client to match client interface signature, since client.NewSession() does not use an interface
type SshClientWrapper struct {
	sshclient *ssh.Client
}

//NewClientWrapper -
func NewClientWrapper(client *ssh.Client) *SshClientWrapper {
	return &SshClientWrapper{
		sshclient: client,
	}
}

//NewSession -
func (c *SshClientWrapper) NewSession() (SSHSession, error) {
	return c.sshclient.NewSession()
}

//NewRemoteExecutor - This method creates executor based on ssh, it has concrete ssh reference
func NewRemoteExecutor(sshCfg SshConfig) (executor Executor, err error) {
	clientconfig := &ssh.ClientConfig{
		User: sshCfg.Username,
		Auth: sshCfg.GetAuthMethod(),
	}
	remoteExecutor := &DefaultRemoteExecutor{}
	remoteExecutor.LazyClientDial = func() {
		log.Printf("sshing to %s:%d with %s:%s", sshCfg.Host, sshCfg.Port, sshCfg.Username, sshCfg.Password)
		client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", sshCfg.Host, sshCfg.Port), clientconfig)
		if err != nil {
			lo.G.Error("ssh connection issue:", err)
			return
		}
		remoteExecutor.Client = NewClientWrapper(client)
		remoteExecutor.sshclient = client
	}
	executor = remoteExecutor
	return
}

//SSHSession -
type SSHSession interface {
	Start(cmd string) error
	Wait() error
	StdoutPipe() (io.Reader, error)
	Close() error
}

func (executor *DefaultRemoteExecutor) SecureCopy(src string, w io.Writer) (err error) {
	if executor.once.Do(executor.LazyClientDial); executor.Client != nil {
		c, err := sftp.NewClient(executor.sshclient, sftp.MaxPacket(DefaultPacketSize))
		if err != nil {
			log.Fatalf("unable to start sftp subsytem: %v", err)
		}
		defer c.Close()

		r, err := c.Open(src)
		if err != nil {
			log.Fatal(err)
		}
		defer r.Close()

		const size int64 = 1e9

		//log.Printf("reading %v bytes", size)
		t1 := time.Now()
		n, err := io.Copy(w, io.LimitReader(r, size))
		if err != nil {
			err = fmt.Errorf("error copying bytes to file %s", src)
		}
		if n != size {
			log.Printf("copy: expected %v bytes, got %d", size, n)
		}
		log.Printf("read %v bytes in %s", size, time.Since(t1))

	} else {
		err = fmt.Errorf("un-initialized client executor")
		lo.G.Error(err.Error())
	}
	return err
}

//Execute - Copy the output from a command to the specified io.Writer
func (executor *DefaultRemoteExecutor) ExecuteForWrite(dest io.Writer, command string) (err error) {
	var session SSHSession
	var stdout io.Reader

	if executor.once.Do(executor.LazyClientDial); executor.Client != nil {
		session, err = executor.Client.NewSession()
		defer session.Close()
		if err != nil {
			return
		}
		stdout, err = session.StdoutPipe()
		if err != nil {
			return
		}
		err = session.Start(command)
		if err != nil {
			return
		}
		_, err = io.Copy(dest, stdout)
		if err != nil {
			return
		}
		err = session.Wait()
	} else {
		err = fmt.Errorf("un-initialized client executor")
		lo.G.Error(err.Error())
	}
	return
}

func (executor *DefaultRemoteExecutor) ExecuteForRead(command string) (stdoutReader io.Reader, err error) {
	var session SSHSession

	if executor.once.Do(executor.LazyClientDial); executor.Client != nil {
		session, err = executor.Client.NewSession()
		defer session.Close()
		if err != nil {
			return
		}
		stdoutReader, err = session.StdoutPipe()
		if err != nil {
			return
		}
		err = session.Start(command)
		if err != nil {
			return
		}
		err = session.Wait()
	} else {
		err = fmt.Errorf("un-initialized client executor")
		lo.G.Error(err.Error())
	}
	return
}
