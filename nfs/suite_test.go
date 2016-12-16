package nfs_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGoblob(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "NFS Suite")
}
