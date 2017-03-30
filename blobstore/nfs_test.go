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

package blobstore_test

import (
	"errors"

	"github.com/pivotalservices/goblob/blobstore"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("NFS", func() {
	var store blobstore.Blobstore
	BeforeEach(func() {
		store = blobstore.NewNFS("nfs_testdata")
	})

	Describe("List()", func() {
		It("Should return a list of blobs", func() {
			blobs, err := store.List()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(len(blobs)).Should(BeEquivalentTo(7))
			Ω(blobs[0].Checksum).Should(BeEquivalentTo("b026324c6904b2a9cb4b88d6d61c81d1"))
			Ω(blobs[0].Path).Should(BeEquivalentTo("cc-buildpacks/ea/07/ea07de9b-dd94-477c-b904-0f77d47dd111_a32d9ae40371d557c7c90eb2affc3d7bba6abe69"))
		})
	})
	Describe("Read()", func() {
		It("Given a file it should return a reader", func() {
			blobs, err := store.List()
			Ω(err).ShouldNot(HaveOccurred())

			reader, err := store.Read(blobs[0])
			Ω(err).ShouldNot(HaveOccurred())
			Ω(reader).ShouldNot(BeNil())
		})
	})
	Describe("Write()", func() {
		It("Should return an error", func() {
			err := errors.New("writing to the NFS store is not supported")
			Ω(store.Write(&blobstore.Blob{}, nil)).Should(BeEquivalentTo(err))
		})
	})
})
