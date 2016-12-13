package goblob_test

import (
	"errors"

	. "github.com/c0-ops/goblob"
	"github.com/c0-ops/goblob/mock"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Migrator", func() {
	Describe("When the source store has no files", func() {
		Describe("Migrate(dst, src)", func() {
			var m *CloudFoundryMigrator
			var src *mock.CloudFoundry
			var dst *mock.S3

			BeforeEach(func() {
				m = &CloudFoundryMigrator{}
				src = &mock.CloudFoundry{}
				dst = &mock.S3{}
			})

			It("Should return an error if the destination store is nil", func() {
				var emptyStore = errors.New("src is an empty store")
				err := m.Migrate(dst, nil)
				Ω(err).Should(Equal(emptyStore))
			})

			It("Should return an error if the destination store is nil", func() {
				var emptyStore = errors.New("dst is an empty store")
				err := m.Migrate(nil, src)
				Ω(err).Should(Equal(emptyStore))
			})

			It("Should return an error if the source store has no files", func() {
				var emptyStore = errors.New("the source store has no files")
				src.ListFn = func() ([]Blob, error) {
					return nil, nil
				}
				err := m.Migrate(dst, src)
				Ω(err).Should(Equal(emptyStore))
			})

			It("Should error if the source store errors on List", func() {
				var testErr = errors.New("test")
				src.ListFn = func() ([]Blob, error) {
					return nil, testErr
				}
				err := m.Migrate(dst, src)
				Ω(err).Should(Equal(testErr))
			})
		})
	})
})
