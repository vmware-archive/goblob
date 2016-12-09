package cc_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCc(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cc Suite")
}
