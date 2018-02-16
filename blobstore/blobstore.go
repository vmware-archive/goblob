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

import "io"

// Blob is a file in a blob store
type Blob struct {
	Checksum string
	Path     string
}

//go:generate counterfeiter . Blobstore

type Blobstore interface {
	//Returns logical name of blobstore (S3, NFS)
	Name() string
	//Returns a list of all the "blobs" in blobstore
	List() ([]*Blob, error)
	//For a given blob will return io.ReadCloser with contents
	Read(src *Blob) (io.ReadCloser, error)
	//Returns md5 checksum for the given blob
	Checksum(src *Blob) (string, error)
	//Writes the blob to the blobstore
	Write(dst *Blob, src io.Reader) error
	//Determins if blob exists
	Exists(*Blob) bool
	//Returns an interator for all the blobs in the given bucket (or folder for NFS)
	NewBucketIterator(string) (BucketIterator, error)
}
