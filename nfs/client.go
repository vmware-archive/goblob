package nfs

import (
	"fmt"
	"io"
	"log"
	"os"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	"code.cloudfoundry.org/lager"
	"github.com/pivotal-customer0/cfblobmigrator/cmd"
	"github.com/pivotal-customer0/cfblobmigrator/tar"
)

type Client interface {
	Get(path string) (content io.Reader, err error)
	Put(path string, content io.ReadCloser, contentLength int64) (err error)
	GetAll(path string) (extractPath string, err error)
}

type nfsClient struct {
	NfsDirectory string
	extractor    tar.CmdExtractor
	Caller       cmd.Executor
	logger       lager.Logger
}

var SshCmdExecutor = cmd.NewRemoteExecutor

func NewNFSClient(username string, password string, ip string, remoteArchivePath string, extractor tar.CmdExtractor, logger lager.Logger) (*nfsClient, error) {
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
		Caller:       remoteExecuter,
		extractor:    extractor,
		NfsDirectory: remoteArchivePath,
		logger:       logger,
	}, nil
}

func (c *nfsClient) Get(blobID string) (io.Reader, error) {
	cmd := fmt.Sprintf("cd %s && cat %s", c.NfsDirectory, blobID)
	log.Printf("fetching blob %s with command: %s\n", blobID, cmd)
	return c.Caller.ExecuteForRead(cmd)
}

func (c *nfsClient) GetAll(destinationPath string) (string, error) {
	src := "/tmp/cc-packages.tgz"
	cmd := fmt.Sprintf("cd %s && tar czf %s %s", "/var/vcap/store/shared", src, "cc-packages")
	log.Printf("compressing blobs with command: %s\n", cmd)
	_, err := c.Caller.ExecuteForRead(cmd)
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Compressing blobs with command %s", cmd)
	}

	targetFile, err := os.OpenFile(destinationPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	//targetFile, err := b.fs.OpenFile(destinationPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Opening file for blobs at %s", destinationPath)
	}
	defer targetFile.Close()

	log.Printf("Saving tarball of blobs to %s\n", destinationPath)
	err = c.Caller.SecureCopy(src, targetFile)
	if err != nil {
		return "", bosherr.WrapError(err, "Failed to download blobs")
	}
	extractPath, exErr := c.extractor.Extract(destinationPath)
	if exErr != nil {
		//cleanUpErr := c.downloader.CleanUp(src)
		//if cleanUpErr != nil {
		//	c.logger.Debug("nfsClientGetAll",
		//		"Failed to clean up downloaded blobs %v", cleanUpErr)
		//}

		return extractPath, bosherr.WrapError(exErr, "Extracting blobs")
	}
	return extractPath, err
}

func (c *nfsClient) Put(path string, content io.ReadCloser, contentLength int64) (err error) {
	return bosherr.WrapErrorf(err, "Put function not yet implemented")
}
