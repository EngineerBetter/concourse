package testflight_test

import (
	uuid "github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("A resource pinned with a version during initial set of the pipeline", func() {
	Context("when a resource is pinned in the pipeline config before the check is run", func() {
		BeforeEach(func() {
			hash, err := uuid.NewV4()
			Expect(err).ToNot(HaveOccurred())

			setAndUnpausePipeline(
				"fixtures/version-resource-in-resource.yml",
				"-v", "hash="+hash.String(),
				"-y", `resource_version_config={"version":"v1"}`,
			)
		})

		It("should be able to check the resource", func() {
			check := fly("check-resource", "-r", inPipeline("some-resource"))
			Expect(check).To(gbytes.Say("some-resource.*succeeded"))
		})

		It("should be able to run a job with the pinned version", func() {
			watch := fly("trigger-job", "-j", inPipeline("some-passing-job"), "-w")
			Expect(watch).To(gbytes.Say("v1"))
		})
	})
})
