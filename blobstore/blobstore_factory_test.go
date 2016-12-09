package blobstore_test

import (
	. "github.com/c0-ops/goblob/blobstore"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/c0-ops/goblob/nfs"
	faketar "github.com/c0-ops/goblob/tar/fakes"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
)

var _ = Describe("BlobstoreFactory", func() {
	var (
		fs               *fakesys.FakeFileSystem
		//runner           *fakesys.FakeCmdRunner
		extractor        faketar.FakeCmdExtractor
		logger           boshlog.Logger
		blobstoreFactory BlobstoreFactory
	)

	BeforeEach(func() {
		fs = fakesys.NewFakeFileSystem()
		logger = boshlog.NewLogger(boshlog.LevelNone)
		extractor = faketar.NewFakeCmdExtractor()

		blobstoreFactory = NewRemoteBlobstoreFactory(fs, logger)
	})

	Describe("NewNFSBlobstore", func() {
		It("returns the blobstore", func() {
			blobstore, err := blobstoreFactory.NewBlobstore("fake-user", "fake-password", "fake-ip", extractor)
			Expect(err).ToNot(HaveOccurred())
			nfsClient, err := nfs.NewNFSClient("fake-user", "fake-password", "fake-ip", extractor, fs, logger)
			Expect(err).ToNot(HaveOccurred())
			expectedBlobstore := NewBlobstore(nfsClient, fs, extractor, logger)
			Expect(blobstore).To(BeEquivalentTo(expectedBlobstore))
		})
	})
})
