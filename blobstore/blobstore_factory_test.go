package blobstore_test

import (
	"code.cloudfoundry.org/lager"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	. "github.com/pivotal-customer0/cfblobmigrator/blobstore"

	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	"github.com/pivotal-customer0/cfblobmigrator/nfs"
)

var _ = Describe("BlobstoreFactory", func() {
	var (
		fs               *fakesys.FakeFileSystem
		logger           lager.Logger
		logBuffer        *gbytes.Buffer
		blobstoreFactory Factory
	)

	BeforeEach(func() {
		fs = fakesys.NewFakeFileSystem()
		logger = lager.NewLogger("logger")
		logBuffer = gbytes.NewBuffer()
		logger.RegisterSink(lager.NewWriterSink(logBuffer, lager.INFO))

		blobstoreFactory = NewRemoteBlobstoreFactory(fs, logger)
	})

	Describe("NewNFSBlobstore", func() {
		It("returns the blobstore", func() {
			blobstore, err := blobstoreFactory.NewRemoteBlobstore("fake-user", "fake-password", "fake-ip", "fake-archive-dir", nil, logger)
			Expect(err).ToNot(HaveOccurred())
			nfsClient, err2 := nfs.NewNFSClient("fake-user", "fake-password", "fake-ip", "fake-archive-dir", nil, logger)
			Expect(err2).ToNot(HaveOccurred())
			expectedBlobstore := NewBlobstore(nfsClient, fs, nil, logger)
			Expect(blobstore).To(Equal(expectedBlobstore))
		})
	})
})
