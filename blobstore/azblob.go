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

package blobstore

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/cheggaaa/pb"
	"github.com/pivotal-cf/goblob/validation"

	"github.com/Azure/azure-storage-blob-go/2018-03-28/azblob"
)

var (
	containers              = []string{"cc-buildpacks", "cc-droplets", "cc-packages", "cc-resources"}
	cloudStorageEnpointsMap = map[string]string{
		"AzureChinaCloud":   "core.chinacloudapi.cn",
		"AzureCloud":        "core.windows.net",
		"AzureGermanCloud":  "core.cloudapi.de",
		"AzureUSGovernment": "core.usgovcloudapi.net",
	}
)

type azblobStore struct {
	serviceURL          *azblob.ServiceURL
	useMultipartUploads bool
	containerMapping    map[string]string
}

func NewAzBlobStore(
	accountName string,
	accountKey string,
	cloudName string,
	buildpacksContainerName string,
	dropletsContainerName string,
	packagesContainerName string,
	resourcesContainerName string,
) Blobstore {
	credential := azblob.NewSharedKeyCredential(accountName, accountKey)
	pipeline := azblob.NewPipeline(credential, azblob.PipelineOptions{})

	primaryURL, _ := url.Parse(
		fmt.Sprintf("https://%s.blob.%s", accountName, cloudStorageEnpointsMap[cloudName]))
	serviceURL := azblob.NewServiceURL(*primaryURL, pipeline)

	return &azblobStore{
		serviceURL: &serviceURL,
		containerMapping: map[string]string{
			"cc-buildpacks": buildpacksContainerName,
			"cc-resources":  resourcesContainerName,
			"cc-droplets":   dropletsContainerName,
			"cc-packages":   packagesContainerName,
		},
	}
}

func (s *azblobStore) Name() string {
	return "azure blob store"
}

func (s *azblobStore) List() ([]*Blob, error) {
	var blobs []*Blob

	containersOnAccount, err := s.listContainers()
	if err != nil {
		return nil, err
	}

	for _, container := range containers {
		containerName := s.destContainerName(container)
		containerUrl := s.serviceURL.NewContainerURL(containerName)

		containerExists := stringInArray(containersOnAccount, containerName)
		if containerExists {
			for marker := (azblob.Marker{}); marker.NotDone(); {
				listBlob, err := containerUrl.ListBlobsFlatSegment(context.Background(),
					marker,
					azblob.ListBlobsSegmentOptions{})
				if err != nil {
					return nil, err
				}

				marker = listBlob.NextMarker

				bar := pb.StartNew(len(listBlob.Segment.BlobItems))
				for _, blobInfo := range listBlob.Segment.BlobItems {
					md5, err := base64.StdEncoding.DecodeString(string(blobInfo.Properties.ContentMD5[:]))
					if err != nil {
						return nil, err
					}
					checksum := hex.EncodeToString(md5)

					blob := &Blob{
						Path:     filepath.Join(container, blobInfo.Name),
						Checksum: checksum,
					}
					blobs = append(blobs, blob)

					bar.Increment()
				}
				bar.FinishPrint(containerName)
			}
		}
	}
	return blobs, nil
}

func (s *azblobStore) Read(src *Blob) (io.ReadCloser, error) {
	containerName := s.containerName(src)
	path := s.path(src)

	containerURL := s.serviceURL.NewContainerURL(containerName)
	blobURL := containerURL.NewBlockBlobURL(path)
	response, err := blobURL.Download(context.Background(),
		0,
		azblob.CountToEnd,
		azblob.BlobAccessConditions{},
		false)
	if err != nil {
		return nil, err
	}

	return response.Body(azblob.RetryReaderOptions{MaxRetryRequests: 3}), nil
}

func (s *azblobStore) Checksum(src *Blob) (string, error) {
	// large blob does not have Content-MD5 in Metadata,
	// otherwise it will be faster to get checksum from metadata.
	// return s.checksumFromMetadata(src)
	rc, err := s.Read(src)
	if err != nil {
		return "", nil
	}
	defer rc.Close()

	return validation.ChecksumReader(rc)
}

func (s *azblobStore) Write(dst *Blob, src io.Reader) error {
	containerName := s.containerName(dst)
	path := s.path(dst)
	if err := s.createContainer(containerName); err != nil {
		return err
	}

	containerURL := s.serviceURL.NewContainerURL(containerName)
	blobURL := containerURL.NewBlockBlobURL(path)

	_, err := azblob.UploadStreamToBlockBlob(context.Background(),
		src,
		blobURL,
		azblob.UploadStreamToBlockBlobOptions{
			BufferSize: 10 * 1024 * 1024, // 10M buffer size
			MaxBuffers: 20,
		})
	if err != nil {
		return err
	}
	return nil
}

func (s *azblobStore) Exists(blob *Blob) bool {
	checksum, err := s.Checksum(blob)
	if err != nil {
		return false
	}

	return checksum == blob.Checksum
}

func (s *azblobStore) NewBucketIterator(containerName string) (BucketIterator, error) {
	destContainerName := s.destContainerName(containerName)
	containerExists, err := s.doesContainerExist(destContainerName)
	if err != nil {
		return nil, err
	}

	if !containerExists {
		return nil, errors.New("bucket does not exist")
	}

	containerUrl := s.serviceURL.NewContainerURL(destContainerName)

	marker := azblob.Marker{}
	listBlob, err := containerUrl.ListBlobsFlatSegment(context.Background(),
		marker,
		azblob.ListBlobsSegmentOptions{})

	if err != nil {
		return nil, err
	}

	if len(listBlob.Segment.BlobItems) == 0 {
		return &s3BucketIterator{}, nil
	}

	doneCh := make(chan struct{})
	blobCh := make(chan *Blob)

	iterator := &s3BucketIterator{
		doneCh: doneCh,
		blobCh: blobCh,
	}

	go func() {
		for marker.NotDone() {
			for _, blobInfo := range listBlob.Segment.BlobItems {
				select {
				case <-doneCh:
					doneCh = nil
					return
				default:
					blobCh <- &Blob{
						Path: filepath.Join(containerName, blobInfo.Name),
					}
				}
			}
			marker = listBlob.NextMarker
		}
		close(blobCh)
	}()

	return iterator, nil
}

// helpers

func (s *azblobStore) destContainerName(container string) string {
	return s.containerMapping[container]
}

func (s *azblobStore) containerName(blob *Blob) string {
	return s.destContainerName(blob.Path[:strings.Index(blob.Path, "/")])
}

func (s *azblobStore) doesContainerExist(containerName string) (bool, error) {
	containers, err := s.listContainers()
	if err != nil {
		return false, err
	}
	return stringInArray(containers, containerName), nil
}

func (s *azblobStore) createContainer(containerName string) error {
	containerURL := s.serviceURL.NewContainerURL(containerName)

	_, err := containerURL.Create(context.Background(),
		azblob.Metadata{},
		azblob.PublicAccessNone)
	if serr, ok := err.(azblob.StorageError); ok {
		switch serr.ServiceCode() {
		case azblob.ServiceCodeContainerAlreadyExists:
		default:
			return err
		}
	}

	return nil
}

func (s *azblobStore) listContainers() ([]string, error) {
	var containers []string
	for marker := (azblob.Marker{}); marker.NotDone(); {
		l, err := s.serviceURL.ListContainersSegment(
			context.Background(),
			marker,
			azblob.ListContainersSegmentOptions{})
		if err != nil {
			return nil, err
		}

		marker = l.NextMarker

		for _, c := range l.ContainerItems {
			containers = append(containers, c.Name)
		}
	}

	return containers, nil
}
func (s *azblobStore) path(blob *Blob) string {
	return blob.Path[(strings.Index(blob.Path, "/") + 1):]
}

func (s *azblobStore) checksumFromMetadata(src *Blob) (string, error) {
	containerName := s.containerName(src)
	path := s.path(src)

	containerURL := s.serviceURL.NewContainerURL(containerName)

	blobURL := containerURL.NewBlobURL(path)
	r, err := blobURL.GetProperties(context.Background(), azblob.BlobAccessConditions{})
	if err != nil {
		return "", err
	}
	md5 := r.ContentMD5()

	return hex.EncodeToString(md5), nil
}

func stringInArray(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}
