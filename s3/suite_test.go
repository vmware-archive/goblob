package s3_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var fakeS3EndpointURL string

func TestS3(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "S3 Suite")
}

var _ = BeforeSuite(func() {
	fakeS3EndpointURL = "127.0.0.1:9000"
})
