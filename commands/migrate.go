package commands

import (
	"fmt"

	"code.cloudfoundry.org/workpool"
	"github.com/pivotalservices/goblob"
	"github.com/pivotalservices/goblob/blobstore"
)

type MigrateCommand struct {
	ConcurrentUploads int      `long:"concurrent-uploads" env:"CONCURRENT_UPLOADS" default:"20"`
	Exclusions        []string `long:"exclude" description:"blobstore directories to exclude, e.g. cc-resources"`

	NFS struct {
		Path string `long:"blobstore-path" env:"BLOBSTORE_PATH" description:"path to root of blobstore" default:"/var/vcap/store/shared"`
	} `group:"NFS"`

	S3 struct {
		AccessKey            string `long:"s3-accesskey" env:"S3_ACCESSKEY" description:"S3 access key"`
		SecretKey            string `long:"s3-secretkey" env:"S3_SECRETKEY" description:"S3 secret access key"`
		Region               string `long:"region" default:"us-east-1" env:"S3_REGION" description:"S3 region"`
		Endpoint             string `long:"s3-endpoint" default:"https://s3.amazonaws.com" env:"S3_ENDPOINT"`
		UseMultipartUploads  bool   `long:"use-multipart-uploads" env:"USE_MULTIPART_UPLOADS"`
		DisableSSL           bool   `long:"disable-ssl" description:"disable SSL connections to S3 endpoint"`
		BuildpacksBucketName string `long:"buildpacks-bucket-name" default:"cc-buildpacks" description:"name of bucket to store buildpacks in"`
		DropletsBucketName   string `long:"droplets-bucket-name" default:"cc-droplets" description:"name of bucket to store droplets in"`
		PackagesBucketName   string `long:"packages-bucket-name" default:"cc-packages" description:"name of bucket to store packages in"`
		ResourcesBucketName  string `long:"resources-bucket-name" default:"cc-resources" description:"name of bucket to store resources in"`
	} `group:"S3"`
}

func (c *MigrateCommand) Execute([]string) error {
	nfsStore := blobstore.NewNFS(c.NFS.Path)
	s3Store := blobstore.NewS3(
		c.S3.AccessKey,
		c.S3.SecretKey,
		c.S3.Region,
		c.S3.Endpoint,
		c.S3.UseMultipartUploads,
		c.S3.DisableSSL,
		c.S3.BuildpacksBucketName,
		c.S3.DropletsBucketName,
		c.S3.PackagesBucketName,
		c.S3.ResourcesBucketName,
	)

	blobMigrator := goblob.NewBlobMigrator(s3Store, nfsStore)
	pool, err := workpool.NewWorkPool(c.ConcurrentUploads)
	if err != nil {
		return fmt.Errorf("error creating workpool: %s", err)
	}

	watcher := goblob.NewBlobstoreMigrationWatcher()

	blobStoreMigrator := goblob.NewBlobstoreMigrator(pool, blobMigrator, c.Exclusions, watcher)

	return blobStoreMigrator.Migrate(s3Store, nfsStore)
}
