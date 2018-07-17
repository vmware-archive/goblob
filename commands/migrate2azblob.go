// Copyright 2017-Present Pivotal Software, Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http:#www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package commands

import (
	"fmt"

	"code.cloudfoundry.org/workpool"
	"github.com/pivotal-cf/goblob"
	"github.com/pivotal-cf/goblob/blobstore"
)

type MigrateToAzureBlobCommand struct {
	ConcurrentUploads int      `long:"concurrent-uploads" env:"CONCURRENT_UPLOADS" default:"20"`
	Exclusions        []string `long:"exclude" description:"blobstore directories to exclude, e.g. cc-resources"`

	NFS struct {
		Path string `long:"blobstore-path" env:"BLOBSTORE_PATH" description:"path to root of blobstore" default:"/var/vcap/store/shared"`
	} `group:"NFS"`

	AzStore struct {
		AccountName          string `long:"azure-storage-account" env:"AZURE_STORAGE_ACCOUNT" description:"Azure storage account name"`
		AccountKey           string `long:"azure-storage-account-key" env:"AZURE_STORAGE_ACCOUNT_KEY" description:"Azure storage account key"`
		CloudName            string `long:"cloud-name" default:"AzureCloud" env:"AZURE_CLOUD" description:"cloud name, available names are: AzureCloud, AzureChinaCloud, AzureGermanCloud, AzureUSGovernment"`
		BuildpacksBucketName string `long:"buildpacks-bucket-name" default:"cc-buildpacks" description:"name of bucket to store buildpacks in"`
		DropletsBucketName   string `long:"droplets-bucket-name" default:"cc-droplets" description:"name of bucket to store droplets in"`
		PackagesBucketName   string `long:"packages-bucket-name" default:"cc-packages" description:"name of bucket to store packages in"`
		ResourcesBucketName  string `long:"resources-bucket-name" default:"cc-resources" description:"name of bucket to store resources in"`
	} `group:"AzureBlob"`
}

func (c *MigrateToAzureBlobCommand) Execute([]string) error {
	nfsStore := blobstore.NewNFS(c.NFS.Path)
	azblobStore := blobstore.NewAzBlobStore(
		c.AzStore.AccountName,
		c.AzStore.AccountKey,
		c.AzStore.CloudName,
		c.AzStore.BuildpacksBucketName,
		c.AzStore.DropletsBucketName,
		c.AzStore.PackagesBucketName,
		c.AzStore.ResourcesBucketName,
	)

	blobMigrator := goblob.NewBlobMigrator(azblobStore, nfsStore)
	pool, err := workpool.NewWorkPool(c.ConcurrentUploads)
	if err != nil {
		return fmt.Errorf("error creating workpool: %s", err)
	}

	watcher := goblob.NewBlobstoreMigrationWatcher()

	blobStoreMigrator := goblob.NewBlobstoreMigrator(pool, blobMigrator, c.Exclusions, watcher)

	return blobStoreMigrator.Migrate(azblobStore, nfsStore)
}
