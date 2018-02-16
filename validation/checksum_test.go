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

package validation_test

import (
	"path"

	. "github.com/pivotal-cf/goblob/validation"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Md5", func() {
	It("Generates correct checksums", func() {
		checksum, err := Checksum(path.Join(".", "fixtures", "testfile"))
		Ω(err).Should(BeNil())
		Ω(checksum).Should(BeEquivalentTo("b026324c6904b2a9cb4b88d6d61c81d1"))
	})

	It("Generates correct checksums", func() {
		checksum, err := Checksum(path.Join(".", "fixtures", "013110a30e2a475551c801b4c45e497ce71c26fe"))
		Ω(err).Should(BeNil())
		Ω(checksum).Should(BeEquivalentTo("9e63a667623321944e174d3d3ea16e9e"))
	})

	It("Returns an error for a missing filename", func() {
		checksum, err := Checksum(path.Join(".", "fixtures", "testmissing"))
		Ω(err).ShouldNot(BeNil())
		Ω(checksum).Should(BeEquivalentTo(""))
	})
})
