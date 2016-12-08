package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	"code.cloudfoundry.org/lager"
	"github.com/pivotal-customer0/cfblobmigrator/blobstore"
	"github.com/pivotal-customer0/cfblobmigrator/s3"
	"github.com/pivotal-customer0/cfblobmigrator/tar"
)

const (
	NfsBlobstorePath string = "/var/vcap/store/shared"
	NfsBuildpacksDir string = "cc-buildpacks/ea/07"
)

var (
	nfsIpAddress = flag.String("host", "localhost", "nfs server ip address")
	vcapPass = flag.String("pass", os.Getenv("VCAP_PASSWORD"), "vcap password for nfs-server job")
	bpBucket = flag.String("buildpacks-bucket", "cc-buildpacks", "S3 bucket for storing app buildpacks. Defaults to cc-buildpacks")
	drpBucket = flag.String("droplets-bucket", "cc-droplets", "S3 bucket for storing app droplets. Defaults to cc-droplets")
	pkgBucket = flag.String("packages-bucket", "cc-packages", "S3 bucket for storing app packages. Defaults to cc-packages")
	resBucket = flag.String("resources-bucket", "cc-resources", "S3 bucket for storing app resources. Defaults to cc-resources")
)

func init() {
	flag.Parse()
}

func main() {
	log := lager.NewLogger("cfblobmigrator-logger")
	log.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))
	log.RegisterSink(lager.NewWriterSink(os.Stderr, lager.ERROR))

	// s3 connection info needs to be moved into a config file
	endpoint := "127.0.0.1:9000"
	accessKeyID := "O1DIN1QXYAN0TBNHHFP7"
	secretAccessKey := "42hF6PdMdDt6IY1sYr91SeGIXJyTCISfFZynmXlZ"
	region := "us-east-1"

	buckets := []string{*bpBucket, *drpBucket, *pkgBucket, *resBucket}

	logger := boshlog.NewWriterLogger(boshlog.LevelDebug, os.Stderr, os.Stderr)
	fs := boshsys.NewOsFileSystem(logger)

	if isLocal() {
		blobstoreFactory := blobstore.NewLocalBlobstoreFactory(fs, log)
		localBlobstore, err := blobstoreFactory.NewBlobstore(log)
		if err != nil {
			log.Fatal("failed to create local blobstore", err)
		}

		for _, bucketName := range buckets {
			if strings.Trim(bucketName, "ignore") != "" {
				destinationPath := "./blobstore/fixtures/"+bucketName
				fmt.Printf("get all blobs from %s\n", destinationPath)
				blobs, err := localBlobstore.GetAll(destinationPath)
				if err != nil {
					log.Fatal("failed to get all blobs", err)
				}
				populateS3blobstore(endpoint, accessKeyID, secretAccessKey, bucketName, region, blobs, log)
			}
		}

		os.Exit(0)
	}

	if *nfsIpAddress == "" {
		fmt.Println("You must specify the nfs ip address")
	}
	for _, bucketName := range buckets {
		if bucketName != "" {
			runner := boshsys.NewExecCmdRunner(logger)
			nfsDirectory := path.Join(NfsBlobstorePath, NfsBuildpacksDir)
			extractor := tar.NewCmdExtractor(runner, fs, logger)
			blobstoreFactory := blobstore.NewRemoteBlobstoreFactory(fs, log)

			nfsBlobstore, err := blobstoreFactory.NewBlobstore("vcap", *vcapPass, *nfsIpAddress, nfsDirectory, extractor, log)

			file, err := fs.TempFile("bosh-init-local-blob")
			destinationPath := file.Name()
			err = file.Close()
			if err != nil {
				log.Fatal("failed closing file", err)
			}
			fmt.Printf("get all blobs from %s\n", destinationPath)
			blobs, err := nfsBlobstore.GetAll(destinationPath)
			if err != nil {
				log.Fatal("failed to get all blobs", err)
			}
			populateS3blobstore(endpoint, accessKeyID, secretAccessKey, bucketName, region, blobs, log)
		}
	}
}

func populateS3blobstore(endpoint, accessKeyID, secretAccessKey, bucketName, region string, blobs []blobstore.LocalBlob, logger lager.Logger) {
	err := createBucket(endpoint, accessKeyID, secretAccessKey, bucketName, region, logger)
	if err != nil {
		logger.Fatal("failed to create bucket", err)
	}
	for _, blob := range blobs {
		fmt.Printf("uploading blob %s\n", blob.Path())
		err = uploadObject(endpoint, accessKeyID, secretAccessKey, bucketName, blob, logger)
		if err != nil {
			logger.Fatal("failed to upload blob %s" + blob.Path(), err)
		}
	}
}

func isLocal() bool {
	if *vcapPass == "" {
		return true
	}
	return false
}

func createBucket(endpoint, accessKeyID, secretAccessKey, bucketName, region string, logger lager.Logger) error {
	client, err := s3.NewClient(endpoint, accessKeyID, secretAccessKey, false, logger)
	if err != nil {
		return err
	}
	err = client.CreateBucket(bucketName, region)
	return err
}

func uploadObject(endpoint, accessKeyID, secretAccessKey, bucketName string, blob blobstore.LocalBlob, logger lager.Logger) error {
	client, err := s3.NewClient(endpoint, accessKeyID, secretAccessKey, false, logger)
	if err != nil {
		return err
	}
	//object, _ := fs.OpenFile(blob.Path(), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	object, fileOpenErr := os.Open(blob.Path())
	if fileOpenErr != nil {
		return fileOpenErr
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
