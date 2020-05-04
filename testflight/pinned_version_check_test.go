package testflight_test

import (
	uuid "github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("A resource is configured with a version during initial set of the pipeline", func() {
	var hash *uuid.UUID
	var err error
	BeforeEach(func() {
		hash, err = uuid.NewV4()
		Expect(err).ToNot(HaveOccurred())
	})

	Context("when a resource is pinned in the pipeline config before the check is run", func() {
		BeforeEach(func() {
			setAndUnpausePipeline(
				"fixtures/version-resource-in-resource.yml",
				"-v", "hash="+hash.String(),
				"-y", `version_config={"version":"v1"}`,
			)
		})

		It("should be able to check the resource", func() {
			check := fly("check-resource", "-r", inPipeline("some-resource"))
			Expect(check).To(gbytes.Say("some-resource.*succeeded"))
		})

		It("should be able to run a job with the pinned version", func() {
			_ = newMockVersion("some-resource", "v1")
			_ = newMockVersion("some-resource", "v2")
			watch := fly("trigger-job", "-j", inPipeline("some-passing-job"), "-w")
			Expect(watch).To(gbytes.Say("v1"))
			Expect(watch).NotTo(gbytes.Say("v2"))
		})
	})

	Context("when a resource is configured to use latest version in the pipeline config before the check is run", func() {
		BeforeEach(func() {
			setAndUnpausePipeline(
				"fixtures/version-resource-in-resource.yml",
				"-v", "hash="+hash.String(),
				"-v", "version_config=latest",
			)
		})

		It("should be able to check the resource", func() {
			check := fly("check-resource", "-r", inPipeline("some-resource"))
			Expect(check).To(gbytes.Say("some-resource.*succeeded"))
		})

		It("should be able to run a job with the latest version", func() {
			_ = newMockVersion("some-resource", "v1")
			_ = newMockVersion("some-resource", "v2")
			watch := fly("trigger-job", "-j", inPipeline("some-passing-job"), "-w")
			Expect(watch).To(gbytes.Say("v2"))
			Expect(watch).NotTo(gbytes.Say("v1"))
		})
	})

	Context("when a resource is configured to use every version in the pipeline config before the check is run", func() {
		BeforeEach(func() {
			setAndUnpausePipeline(
				"fixtures/version-resource-in-resource.yml",
				"-v", "hash="+hash.String(),
				"-v", "version_config=every",
			)
		})

		It("should be able to check the resource", func() {
			check := fly("check-resource", "-r", inPipeline("some-resource"))
			Expect(check).To(gbytes.Say("some-resource.*succeeded"))
		})

		It("runs builds with every version", func() {
			_ = newMockVersion("some-resource", "v1")
			watch := fly("trigger-job", "-j", inPipeline("some-passing-job"), "-w")
			Expect(watch).To(gbytes.Say("v1"))

			_ = newMockVersion("some-resource", "v2")
			_ = newMockVersion("some-resource", "v3")

			watch = fly("trigger-job", "-j", inPipeline("some-passing-job"), "-w")
			Expect(watch).To(gbytes.Say("v2"))

			watch = fly("trigger-job", "-j", inPipeline("some-passing-job"), "-w")
			Expect(watch).To(gbytes.Say("v3"))
		})
	})
})
