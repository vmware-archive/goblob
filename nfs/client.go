package nfs

import (
	"fmt"
	"io"
	"log"

	boshsys "github.com/cloudfoundry/bosh-utils/system"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	"code.cloudfoundry.org/lager"
	"github.com/pivotal-customer0/cfblobmigrator/cmd"
	"github.com/pivotal-customer0/cfblobmigrator/tar"
	"path"
)

type Client interface {
	Get(path string) (content io.Reader, err error)
	Put(path string, content io.ReadCloser, contentLength int64) (err error)
	GetAll(path string) (extractPath string, err error)
}

type nfsClient struct {
	NfsDirectory string
	fs        boshsys.FileSystem
	extractor    tar.CmdExtractor
	Caller       cmd.Executor
	logger       lager.Logger
}

var SshCmdExecutor = cmd.NewRemoteExecutor

func NewNFSClient(username string, password string, ip string, remoteArchivePath string, extractor tar.CmdExtractor, fs boshsys.FileSystem, logger lager.Logger) (*nfsClient, error) {
	config := cmd.SshConfig{
		Username: username,
		Password: password,
		Host:     ip,
		Port:     22,
	}
	remoteExecuter, err := SshCmdExecutor(config)
	if err != nil {
		return nil, err
	}
	return &nfsClient{
		NfsDirectory: remoteArchivePath,
		fs: fs,
		Caller:       remoteExecuter,
		extractor:    extractor,
		logger:       logger,
	}, nil
}

func (c *nfsClient) Get(blobID string) (io.Reader, error) {
	cmd := fmt.Sprintf("cd %s && cat %s", c.NfsDirectory, blobID)
	log.Printf("fetching blob %s with command: %s\n", blobID, cmd)
	return c.Caller.ExecuteForRead(cmd)
}

func (c *nfsClient) GetAll(blobPath string) (string, error) {
	src := path.Join("/tmp", blobPath) + ".tgz"
	cmd := fmt.Sprintf("cd %s && tar czf %s %s", "/var/vcap/store/shared", src, blobPath)
	log.Printf("compressing blobs with command: %s\n", cmd)
	_, err := c.Caller.ExecuteForRead(cmd)
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Compressing blobs with command %s", cmd)
	}

	tmpFile, err := c.fs.TempFile("bosh-local-blob")
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Opening file for blobs")
	}
	defer tmpFile.Close()
	tmpDir := tmpFile.Name()

	log.Printf("Downloading tarball of blobs to %s\n", tmpDir)
	err = c.Caller.SecureCopy(src, tmpFile)
	if err != nil {
		return "", bosherr.WrapError(err, "Failed to download blobs")
	}
	extractPath, err := c.extractor.Extract(tmpDir)
	defer c.cleanup(tmpDir)
	if err != nil {
		return extractPath, bosherr.WrapError(err, "Extracting blobs")
	}
	return extractPath, err
}

func (c *nfsClient) Put(path string, content io.ReadCloser, contentLength int64) (err error) {
	return bosherr.WrapErrorf(err, "Put function not yet implemented")
}

func (c *nfsClient) cleanup(path string)  {
	err := c.extractor.CleanUp(path)
	if err != nil {
		c.logger.Debug("Failed to remove blobstore tarball %s" + err.Error())
	}
}
