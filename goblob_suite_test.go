package goblob_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGoblob(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Goblob Suite")
}
