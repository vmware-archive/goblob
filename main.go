package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	"github.com/c0-ops/goblob/blobstore"
	"github.com/c0-ops/goblob/s3"
	"github.com/c0-ops/goblob/tar"
)

const (
	NfsBlobstorePath string = "/var/vcap/store/shared"
	NfsBuildpacksDir string = "cc-buildpacks/ea/07"
)

const goblobMainLogTag = "CmdExtractor"

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

	if isLocal() {
		blobstoreFactory := blobstore.NewLocalBlobstoreFactory(fs, logger)
		localBlobstore, err := blobstoreFactory.NewBlobstore(logger)
		if err != nil {
			logger.Error(goblobMainLogTag, "Failed to create local blobstore %v", err)
			os.Exit(1)
		}

		for _, bucketName := range buckets {
			if strings.Trim(bucketName, "ignore") != "" {
				destinationPath := "./blobstore/fixtures/"+bucketName
				fmt.Printf("get all blobs from %s\n", destinationPath)
				blobs, err := localBlobstore.GetAll(destinationPath)
				if err != nil {
					logger.Error(goblobMainLogTag, "Failed to get all blobs %v", err)
					os.Exit(1)
				}
				populateS3blobstore(endpoint, accessKeyID, secretAccessKey, bucketName, region, blobs, logger)
			}
		}
		os.Exit(0)
	}

	if *nfsIpAddress == "" {
		fmt.Println("You must specify the nfs ip address")
	}
	for _, bucketName := range buckets {
		if strings.Trim(bucketName, "ignore") != "" {
			runner := boshsys.NewExecCmdRunner(logger)
			nfsDirectory := path.Join(NfsBlobstorePath, NfsBuildpacksDir)
			extractor := tar.NewCmdExtractor(runner, fs, logger)
			blobstoreFactory := blobstore.NewRemoteBlobstoreFactory(fs, logger)

			nfsBlobstore, err := blobstoreFactory.NewBlobstore("vcap", *vcapPass, *nfsIpAddress, nfsDirectory, extractor, logger)

			blobs, err := nfsBlobstore.GetAll(bucketName)
			if err != nil {
				logger.Error(goblobMainLogTag, "Failed to get all blobs %v", err)
				os.Exit(1)
			}
			populateS3blobstore(endpoint, accessKeyID, secretAccessKey, bucketName, region, blobs, logger)
		}
	}
}

func populateS3blobstore(endpoint, accessKeyID, secretAccessKey, bucketName, region string, blobs []blobstore.LocalBlob, logger boshlog.Logger) {
	err := createBucket(endpoint, accessKeyID, secretAccessKey, bucketName, region, logger)
	if err != nil {
		logger.Error(goblobMainLogTag, "Failed to create bucket %v", err)
	}
	for _, blob := range blobs {
		fmt.Printf("uploading blob %s\n", blob.Path())
		err = uploadObject(endpoint, accessKeyID, secretAccessKey, bucketName, blob, logger)
		if err != nil {
			logger.Error(goblobMainLogTag, "Failed to upload blobs %s %v", blob.Path())
		}
	}
}

func isLocal() bool {
	if *vcapPass == "" {
		return true
	}
	return false
}

func createBucket(endpoint, accessKeyID, secretAccessKey, bucketName, region string, logger boshlog.Logger) error {
	client, err := s3.NewClient(endpoint, accessKeyID, secretAccessKey, false, logger)
	if err != nil {
		return err
	}
	err = client.CreateBucket(bucketName, region)
	return err
}

func uploadObject(endpoint, accessKeyID, secretAccessKey, bucketName string, blob blobstore.LocalBlob, logger boshlog.Logger) error {
	client, err := s3.NewClient(endpoint, accessKeyID, secretAccessKey, false, logger)
	if err != nil {
		return err
	}
	//object, _ := fs.OpenFile(blob.Path(), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	object, err := os.Open(blob.Path())
	if err != nil {
		return err
	}

	dir, file := split(blob.Path())
	fmt.Printf("dir is %s\n", dir)
	fmt.Printf("file is %s\n", file)

	_, err = client.UploadObject(bucketName, file, object, "")
	return err
}

func split(path string) (dir, file string) {
	i := strings.LastIndex(path, "/")
	return path[:i+1], path[i+1:]
}
