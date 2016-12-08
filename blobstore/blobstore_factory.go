package blobstore

import (
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	"code.cloudfoundry.org/lager"
	"github.com/pivotal-customer0/cfblobmigrator/nfs"
	"github.com/pivotal-customer0/cfblobmigrator/tar"
)

type BlobstoreFactory interface {
	NewBlobstore(username string, password string, ip string, remoteArchivePath string, extractor tar.CmdExtractor, logger lager.Logger) (Blobstore, error)
}

type blobstoreFactory struct {
	fs     boshsys.FileSystem
	logger lager.Logger
}

func NewRemoteBlobstoreFactory(fs boshsys.FileSystem, logger lager.Logger) BlobstoreFactory {
	return blobstoreFactory{
		fs:     fs,
		logger: logger,
	}
}

func (f blobstoreFactory) NewBlobstore(username string, password string, ip string, remoteArchivePath string, extractor tar.CmdExtractor, logger lager.Logger) (Blobstore, error) {

	nfsClient, err := nfs.NewNFSClient(username, password, ip, remoteArchivePath, extractor, logger)
	if err != nil {
		return nil, err
	}

	return NewBlobstore(nfsClient, f.fs, extractor, f.logger), nil
}