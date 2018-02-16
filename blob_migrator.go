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

package goblob

import (
	"fmt"

	"github.com/pivotal-cf/goblob/blobstore"
)

//go:generate counterfeiter . BlobMigrator

type BlobMigrator interface {
	Migrate(blob *blobstore.Blob) error
}

type blobMigrator struct {
	dst blobstore.Blobstore
	src blobstore.Blobstore
}

func NewBlobMigrator(dst blobstore.Blobstore, src blobstore.Blobstore) BlobMigrator {
	return &blobMigrator{
		dst: dst,
		src: src,
	}
}

func (m *blobMigrator) Migrate(blob *blobstore.Blob) error {
	reader, err := m.src.Read(blob)
	if err != nil {
		return fmt.Errorf("error reading blob at %s: %s", blob.Path, err)
	}
	defer reader.Close()

	err = m.dst.Write(blob, reader)
	if err != nil {
		return fmt.Errorf("error writing blob at %s: %s", blob.Path, err)
	}

	checksum, err := m.dst.Checksum(blob)
	if err != nil {
		return fmt.Errorf("error checksumming blob at %s: %s", blob.Path, err)
	}

	if checksum != blob.Checksum {
		return fmt.Errorf(
			"error at %s: checksum [%s] does not match [%s]",
			blob.Path,
			checksum,
			blob.Checksum,
		)
	}

	return nil
}
