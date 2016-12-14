package goblob_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path"

	. "github.com/c0-ops/goblob"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/c0-ops/goblob/ssh/fakes"
	"github.com/xchapter7x/lo"
)

var _ = Describe("RemoteNFS", func() {
	var remoteNFS Store
	var fakeExecutor *fakes.FakeExecutor
	var pwd, _ = os.Getwd()
	var tmpDirName = "temp"
	var tmpPath = path.Join(pwd, tmpDirName)
	var controlOutputDir string

	BeforeEach(func() {
		os.MkdirAll(tmpPath, 0700)
		controlOutputDir, _ = ioutil.TempDir(tmpPath, "blobs")
		fakeExecutor = &fakes.FakeExecutor{}
		remoteNFS = NewRemoteNFS(fakeExecutor, controlOutputDir)
	})
	AfterEach(func() {
		lo.G.Debug("removing: ", controlOutputDir)
		os.RemoveAll(tmpPath)
	})

	Describe("List()", func() {
		It("Should return a list of blobs", func() {
			reader, err := os.Open("./fixtures/blobs.tar.gz")
			Ω(err).ShouldNot(HaveOccurred())
			fakeExecutor.ExecuteForReadReturns(reader, nil)
			blobs, err := remoteNFS.List()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(len(blobs)).Should(BeEquivalentTo(7))
			Ω(fakeExecutor.ExecuteForReadCallCount()).Should(BeEquivalentTo(1))
			cmdExecuted := fakeExecutor.ExecuteForReadArgsForCall(0)
			Ω(cmdExecuted).Should(BeEquivalentTo("cd /var/vcap/store/shared && tar -cz ."))
		})
	})
	Describe("Read()", func() {
		It("Given a file it should return a reader", func() {
			reader, err := os.Open("./fixtures/blobs.tar.gz")
			Ω(err).ShouldNot(HaveOccurred())
			fakeExecutor.ExecuteForReadReturns(reader, nil)
			blobs, err := remoteNFS.List()
			Ω(err).ShouldNot(HaveOccurred())

			theReader, theError := remoteNFS.Read(blobs[0])
			Ω(theError).ShouldNot(HaveOccurred())
			Ω(theReader).ShouldNot(BeNil())
		})
	})
	Describe("Write()", func() {
		It("Should return an error", func() {
			err := errors.New("not implemented")
			Ω(remoteNFS.Write(Blob{}, nil)).Should(BeEquivalentTo(err))
		})
	})
})
