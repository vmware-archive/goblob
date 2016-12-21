package main

import (
	"errors"
	"os"

	"github.com/c0-ops/goblob"
	"github.com/c0-ops/goblob/nfs"
	"github.com/c0-ops/goblob/s3"
	"github.com/urfave/cli"
	"github.com/xchapter7x/lo"
)

var (
	VERSION = ""
)

const mainLogTag = "main"

//ErrorHandler -
type ErrorHandler struct {
	ExitCode int
	Error    error
}

func main() {
	eh := new(ErrorHandler)
	eh.ExitCode = 0
	app := NewApp(eh)
	if err := app.Run(os.Args); err != nil {
		eh.ExitCode = 1
		eh.Error = err
		lo.G.Error(eh.Error)
	}
	os.Exit(eh.ExitCode)
}

// NewApp creates a new cli app
func NewApp(eh *ErrorHandler) *cli.App {
	//cli.AppHelpTemplate = CfopsHelpTemplate
	app := cli.NewApp()
	app.Version = VERSION
	app.Name = "goblob"
	app.Usage = "goblob"
	app.Commands = []cli.Command{
		cli.Command{
			Name:  "version",
			Usage: "shows the application version currently in use",
			Action: func(c *cli.Context) (err error) {
				cli.ShowVersion(c)
				return
			},
		},
		CreateMigrateNFSCommand(eh),
	}
	return app
}

func CreateMigrateNFSCommand(eh *ErrorHandler) cli.Command {
	return cli.Command{
		Action:      nfsAction,
		Name:        "migrate",
		Usage:       "migrate nfs blobs to s3",
		Description: "migrate nfs blobs to s3",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "blobstore-path", Value: "/var/vcap/store/shared", Usage: "path to root of blobstore", EnvVar: "BLOBSTORE_PATH"},
			cli.StringFlag{Name: "cf-identifier", Value: "", Usage: "unique identifier for cloud foundary deployment", EnvVar: "CF_IDENTIFIER"},
			cli.StringFlag{Name: "s3-accesskey", Value: "", Usage: "s3 access key", EnvVar: "S3_ACCESSKEY"},
			cli.StringFlag{Name: "s3-secretkey", Value: "", Usage: "s3 secrety key", EnvVar: "S3_SECRETKEY"},
			cli.StringFlag{Name: "s3-region", Value: "us-east-1", Usage: "s3 region", EnvVar: "S3_REGION"},
			cli.StringFlag{Name: "s3-endpoint", Value: "https://s3.amazonaws.com", Usage: "s3 endpoint", EnvVar: "S3_ENDPOINT"},
			cli.IntFlag{Name: "concurrent-uploads", Value: 20, Usage: "number of concurrent uploads", EnvVar: "CONCURRENT_UPLOADS"},
			cli.BoolFlag{Name: "use-multipart-uploads", Usage: "use multi-part uploads", EnvVar: "USE_MULTIPART_UPLOADS"},
		},
	}
}

func nfsAction(c *cli.Context) error {
	cfIdentifier := c.String("cf-identifier")
	awsAccessKey := c.String("s3-accesskey")
	awsSecretKey := c.String("s3-secretkey")

	if cfIdentifier == "" {
		return errors.New("Must provide cf-identifier")
	}
	if awsAccessKey == "" {
		return errors.New("Must provide s3-accesskey")
	}

	if awsSecretKey == "" {
		return errors.New("Must provide s3-secretkey")
	}

	migrator := goblob.New(c.Int("concurrent-uploads"))
	srcStore := nfs.New(c.String("blobstore-path"))
	dstStore := s3.New(c.String("cf-identifier"),
		awsAccessKey,
		awsSecretKey,
		c.String("s3-region"),
		c.String("s3-endpoint"),
		c.Bool("use-multipart-uploads"))
	return migrator.Migrate(dstStore, srcStore)
}
