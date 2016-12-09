package main

import (
	"flag"
	"fmt"
	"os"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	"github.com/c0-ops/goblob/xfer"
	"github.com/c0-ops/goblob/blobstore"
	"github.com/c0-ops/goblob/tar"
)

const mainLogTag = "main"

var (
	nfsIpAddress = flag.String("host", "localhost", "nfs server ip address")
	vcapPass = flag.String("pass", os.Getenv("VCAP_PASSWORD"), "vcap password for nfs-server job")
	bpBucket = flag.String("buildpacks", "cc-buildpacks", "S3 bucket for storing app buildpacks. Defaults to cc-buildpacks")
	drpBucket = flag.String("droplets", "cc-droplets", "S3 bucket for storing app droplets. Defaults to cc-droplets")
	pkgBucket = flag.String("packages", "cc-packages", "S3 bucket for storing app packages. Defaults to cc-packages")
	resBucket = flag.String("resources", "cc-resources", "S3 bucket for storing app resources. Defaults to cc-resources")
)

func init() {
	flag.Parse()
}

func main() {

	// s3 connection info needs to be moved into a config file
	endpoint := "127.0.0.1:9000"
	accessKeyID := "D2Z5WU2UI35D05WXSJGW"
	secretAccessKey := "Y+4XHK07GQbDqQbkVFIgz2VVi3fapWIGfsdpIL0q"
	region := "us-east-1"

	buckets := []string{*bpBucket, *drpBucket, *pkgBucket, *resBucket}

	logger := boshlog.NewWriterLogger(boshlog.LevelDebug, os.Stderr, os.Stderr)
	fs := boshsys.NewOsFileSystem(logger)

	localBlobstoreFactory := blobstore.NewLocalBlobstoreFactory(fs, logger)
	localBlobstore, err := localBlobstoreFactory.NewBlobstore()
	if err != nil {
		logger.Error(mainLogTag, "Failed to create local blobstore %v", err)
		os.Exit(1)
	}

	svc := xfer.NewTransferService(endpoint, accessKeyID, secretAccessKey, region, localBlobstore, logger)

	if isLocal() {
		err := svc.Transfer(buckets, "./blobstore/fixtures")
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}

	if *nfsIpAddress == "" {
		fmt.Println("You must specify the nfs ip address")
		os.Exit(1)
	}

	runner := boshsys.NewExecCmdRunner(logger)
	extractor := tar.NewCmdExtractor(runner, fs, logger)
	blobstoreFactory := blobstore.NewRemoteBlobstoreFactory(fs, logger)

	nfsBlobstore, err := blobstoreFactory.NewBlobstore("vcap", *vcapPass, *nfsIpAddress, extractor)
	if err != nil {
		logger.Error(mainLogTag, "Failed to create nfs blobstore %v", err)
		os.Exit(1)
	}

	rxfer := xfer.NewRemoteTransferService(svc, endpoint, accessKeyID, secretAccessKey, region, nfsBlobstore, logger)
	err = rxfer.Transfer(buckets, "")
	if err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func isLocal() bool {
	if *vcapPass == "" {
		return true
	}
	return false
}

