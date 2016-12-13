package main

func main() {}

//
// import (
// 	"errors"
// 	"flag"
// 	"fmt"
// 	"os"
// 	"time"
//
// 	boshlog "github.com/cloudfoundry/bosh-utils/logger"
// 	boshsys "github.com/cloudfoundry/bosh-utils/system"
//
// 	"github.com/c0-ops/goblob/blobstore"
// 	"github.com/c0-ops/goblob/bosh"
// 	"github.com/c0-ops/goblob/cc"
// 	"github.com/c0-ops/goblob/s3"
// 	"github.com/c0-ops/goblob/tar"
// 	"github.com/c0-ops/goblob/xfer"
// 	"github.com/urfave/cli"
// 	"github.com/xchapter7x/lo"
// )
//
// const mainLogTag = "main"
//
// var (
// 	//VERSION - Application version injected by build
// 	VERSION            string
// 	nfsIPAddress       = flag.String("host", "localhost", "nfs server ip address")
// 	vcapPass           = flag.String("pass", os.Getenv("VCAP_PASSWORD"), "vcap password for nfs-server job")
// 	bpBucket           = flag.String("buildpacks", "cc-buildpacks", "S3 bucket for storing app buildpacks. Defaults to cc-buildpacks")
// 	drpBucket          = flag.String("droplets", "cc-droplets", "S3 bucket for storing app droplets. Defaults to cc-droplets")
// 	pkgBucket          = flag.String("packages", "cc-packages", "S3 bucket for storing app packages. Defaults to cc-packages")
// 	resBucket          = flag.String("resources", "cc-resources", "S3 bucket for storing app resources. Defaults to cc-resources")
// 	endpoint           = flag.String("s3-endpoint", "", "s3 endpoint to migrate blobs to")
// 	accessKeyID        = flag.String("s3-accesskey", os.Getenv("S3_ACCESSKEY"), "access key for s3 blob store, can use S3_ACCESSKEY env variable")
// 	secretAccessKey    = flag.String("s3-secret-key", os.Getenv("S3_SECRETKEY"), "secret key for s3 blob store, can use S3_SECRETKEY env variable")
// 	region             = flag.String("s3-region", "us-east-1", "s3 region for blobstore")
// 	secure             = false
// 	boshDirectorURL    = flag.String("bosh-url", "", "url of bosh director")
// 	boshDirectorUser   = flag.String("bosh-user", "director", "bosh director user id")
// 	boshDirectorPass   = flag.String("bosh-password", "", "bosh director password")
// 	boshDirectorSecure = false
// 	boshDeployment     = flag.String("bosh-deployment", "", "bosh cf deployment")
// 	//localBlobPath = flag.String("path", "./blobstore/fixtures", "path to local blobstore")
// )
//
// //ErrorHandler -
// type ErrorHandler struct {
// 	ExitCode int
// 	Error    error
// }
//
// func main() {
// 	eh := new(ErrorHandler)
// 	eh.ExitCode = 0
// 	app := NewApp(eh)
// 	if err := app.Run(os.Args); err != nil {
// 		eh.ExitCode = 1
// 		eh.Error = err
// 		lo.G.Error(eh.Error)
// 	}
// 	os.Exit(eh.ExitCode)
// }
//
// // NewApp creates a new cli app
// func NewApp(eh *ErrorHandler) *cli.App {
// 	//cli.AppHelpTemplate = CfopsHelpTemplate
// 	app := cli.NewApp()
// 	app.Version = VERSION
// 	app.Name = "goblob"
// 	app.Usage = "goblob"
// 	app.Commands = []cli.Command{
// 		cli.Command{
// 			Name:  "version",
// 			Usage: "shows the application version currently in use",
// 			Action: func(c *cli.Context) (err error) {
// 				cli.ShowVersion(c)
// 				return
// 			},
// 		},
// 		CreateStartCloudControlCommand(eh),
// 		CreateStopCloudControlCommand(eh),
// 		CreateMigrateNFSCommand(eh),
// 	}
// 	return app
// }
//
// func CreateMigrateNFSCommand(eh *ErrorHandler) cli.Command {
// 	return cli.Command{
// 		Action:      nfsAction,
// 		Name:        "migrate",
// 		Usage:       "migrate nfs blobs to s3",
// 		Description: "migrate nfs blobs to s3",
// 		Flags: []cli.Flag{
// 			cli.StringFlag{Name: "bosh-url", Value: "", Usage: "url of bosh director", EnvVar: "BOSH_URL"},
// 		},
// 	}
// }
//
// func nfsAction(c *cli.Context) error {
//
// 	return nil
// }
//
// func CreateStartCloudControlCommand(eh *ErrorHandler) (command cli.Command) {
// 	command = cli.Command{
// 		Action:      ccStart,
// 		Name:        "start-jobs",
// 		Usage:       "starts the cloud controller(s)",
// 		Description: "starts the cloud controller(s)",
// 		Flags: []cli.Flag{
// 			cli.StringFlag{Name: "bosh-url", Value: "", Usage: "url of bosh director", EnvVar: "BOSH_URL"},
// 			cli.StringFlag{Name: "bosh-user", Value: "director", Usage: "user of bosh director", EnvVar: "BOSH_USER"},
// 			cli.StringFlag{Name: "bosh-password", Value: "", Usage: "password of bosh director", EnvVar: "BOSH_PASSWORD"},
// 			cli.StringFlag{Name: "bosh-deployment", Value: "", Usage: "cf deployment name", EnvVar: "BOSH_DEPLOYMENT"},
// 		},
// 	}
// 	return
// }
//
// func CreateStopCloudControlCommand(eh *ErrorHandler) (command cli.Command) {
// 	command = cli.Command{
// 		Action:      ccStop,
// 		Name:        "stop-jobs",
// 		Usage:       "stops the cloud controller(s)",
// 		Description: "stops the cloud controller(s)",
// 		Flags: []cli.Flag{
// 			cli.StringFlag{Name: "bosh-url", Value: "", Usage: "url of bosh director", EnvVar: "BOSH_URL"},
// 			cli.StringFlag{Name: "bosh-user", Value: "director", Usage: "user of bosh director", EnvVar: "BOSH_USER"},
// 			cli.StringFlag{Name: "bosh-password", Value: "", Usage: "password of bosh director", EnvVar: "BOSH_PASSWORD"},
// 			cli.StringFlag{Name: "bosh-deployment", Value: "", Usage: "cf deployment name", EnvVar: "BOSH_DEPLOYMENT"},
// 		},
// 	}
// 	return
// }
//
// func ccStart(c *cli.Context) error {
// 	if cc, err := getCloudController(c); err != nil {
// 		return err
// 	} else {
// 		return cc.Start()
// 	}
// }
//
// func ccStop(c *cli.Context) error {
// 	if cc, err := getCloudController(c); err != nil {
// 		return err
// 	} else {
// 		return cc.Stop()
// 	}
// }
//
// func getCloudController(c *cli.Context) (*cc.CloudController, error) {
// 	boshURL := c.String("bosh-url")
// 	boshUser := c.String("bosh-user")
// 	boshPass := c.String("bosh-password")
// 	boshDeployment := c.String("bosh-deployment")
//
// 	if boshURL == "" || boshUser == "" || boshPass == "" || boshDeployment == "" {
// 		return nil, errors.New("Must supply bosh-url, bosh-user, bosh-password and bosh-deployment flags")
// 	}
//
// 	taskPingFreq := 1000 * time.Millisecond
// 	bc := bosh.NewClient(bosh.Config{
// 		URL:                 boshURL,
// 		Username:            boshUser,
// 		Password:            boshPass,
// 		TaskPollingInterval: taskPingFreq,
// 		AllowInsecureSSL:    true,
// 	})
//
// 	if vms, err := bc.GetVMs(boshDeployment); err != nil {
// 		return nil, err
// 	} else {
// 		return cc.NewCloudController(bc, boshDeployment, vms), nil
// 	}
// }
//
// func main_old() {
// 	buckets := []string{*bpBucket, *drpBucket, *pkgBucket, *resBucket}
//
// 	logger := boshlog.NewWriterLogger(boshlog.LevelDebug, os.Stderr, os.Stderr)
// 	fs := boshsys.NewOsFileSystem(logger)
//
// 	localBlobstoreFactory := blobstore.NewLocalBlobstoreFactory(fs, logger)
// 	localBlobstore, err := localBlobstoreFactory.NewBlobstore()
// 	if err != nil {
// 		logger.Error(mainLogTag, "Failed to create local blobstore %v", err)
// 		os.Exit(1)
// 	}
//
// 	taskPingFreq := 1000 * time.Millisecond
// 	bc := bosh.NewClient(bosh.Config{
// 		URL:                 *boshDirectorURL,
// 		Username:            *boshDirectorUser,
// 		Password:            *boshDirectorPass,
// 		TaskPollingInterval: taskPingFreq,
// 		AllowInsecureSSL:    !boshDirectorSecure,
// 	})
//
// 	vms, err := bc.GetVMs(*boshDeployment)
//
// 	cloudController := cc.NewCloudController(bc, *boshDeployment, vms)
// 	cloudController.Stop()
// 	defer cloudController.Start()
//
// 	s3Client, err := s3.NewClient(
// 		s3.Config{
// 			Endpoint:        *endpoint,
// 			AccessKeyID:     *accessKeyID,
// 			SecretAccessKey: *secretAccessKey,
// 			Region:          *region,
// 			UseSSL:          secure,
// 		}, logger)
// 	if err != nil {
// 		logger.Error(mainLogTag, "Failed to create s3 client %v", err)
// 		os.Exit(1)
// 	}
//
// 	svc := xfer.NewTransferService(s3Client, localBlobstore, logger)
//
// 	if isLocal() {
// 		err = svc.Transfer(buckets, "./blobstore/fixtures")
// 		if err != nil {
// 			os.Exit(1)
// 		}
// 		os.Exit(0)
// 	}
//
// 	if *nfsIPAddress == "" {
// 		fmt.Println("You must specify the nfs ip address")
// 		os.Exit(1)
// 	}
//
// 	runner := boshsys.NewExecCmdRunner(logger)
// 	extractor := tar.NewCmdExtractor(runner, fs, logger)
// 	blobstoreFactory := blobstore.NewRemoteBlobstoreFactory(fs, logger)
// 	sshPort := 2222
// 	nfsBlobstore, err := blobstoreFactory.NewBlobstore("vcap", *vcapPass, *nfsIPAddress, sshPort, extractor)
// 	if err != nil {
// 		logger.Error(mainLogTag, "Failed to create nfs blobstore %v", err)
// 		os.Exit(1)
// 	}
//
// 	rxfer := xfer.NewRemoteTransferService(svc, s3Client, nfsBlobstore, logger)
// 	err = rxfer.Transfer(buckets, "")
// 	if err != nil {
// 		os.Exit(1)
// 	}
// 	os.Exit(0)
// }
//
// func isLocal() bool {
// 	if *vcapPass == "" {
// 		return true
// 	}
// 	return false
// }
