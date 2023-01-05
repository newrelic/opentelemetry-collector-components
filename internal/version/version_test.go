//go:build unit

package version

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Version", func() {
	It("should not be empty", func() {
		Expect(Version).NotTo(Equal(""))
	})
})
