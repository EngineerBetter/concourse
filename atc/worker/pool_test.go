package worker_test

import (
	"code.cloudfoundry.org/lager/lagertest"

	"context"
	"errors"

	"github.com/concourse/baggageclaim"
	"github.com/concourse/concourse/atc"
	"github.com/concourse/concourse/atc/db"
	"github.com/concourse/concourse/atc/db/dbfakes"
	"github.com/concourse/concourse/atc/db/lock/lockfakes"
	"github.com/concourse/concourse/atc/policy"
	"github.com/concourse/concourse/atc/policy/policyfakes"
	. "github.com/concourse/concourse/atc/worker"
	"github.com/concourse/concourse/atc/worker/workerfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Pool", func() {
	var (
		logger            *lagertest.TestLogger
		pool              Pool
		fakeProvider      *workerfakes.FakeWorkerProvider
		fakePolicyChecker *policyfakes.FakeChecker
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("test")
		fakeProvider = new(workerfakes.FakeWorkerProvider)
		fakePolicyChecker = new(policyfakes.FakeChecker)
		fakeLockFactory := new(lockfakes.FakeLockFactory)

		fakeLock := new(lockfakes.FakeLock)
		fakeLockFactory.AcquireReturns(fakeLock, true, nil)
		pool = NewPool(fakeProvider, fakeLockFactory, fakePolicyChecker)
	})

	Describe("FindContainer", func() {
		var (
			foundContainer Container
			found          bool
			findErr        error
		)

		JustBeforeEach(func() {
			foundContainer, found, findErr = pool.FindContainer(
				logger,
				4567,
				"some-handle",
			)
		})

		Context("when looking up the worker errors", func() {
			BeforeEach(func() {
				fakeProvider.FindWorkerForContainerReturns(nil, false, errors.New("nope"))
			})

			It("errors", func() {
				Expect(findErr).To(HaveOccurred())
			})
		})

		Context("when worker is not found", func() {
			BeforeEach(func() {
				fakeProvider.FindWorkerForContainerReturns(nil, false, nil)
			})

			It("returns not found", func() {
				Expect(findErr).NotTo(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})

		Context("when a worker is found with the container", func() {
			var fakeWorker *workerfakes.FakeWorker
			var fakeContainer *workerfakes.FakeContainer

			BeforeEach(func() {
				fakeWorker = new(workerfakes.FakeWorker)
				fakeProvider.FindWorkerForContainerReturns(fakeWorker, true, nil)

				fakeContainer = new(workerfakes.FakeContainer)
				fakeWorker.FindContainerByHandleReturns(fakeContainer, true, nil)
			})

			It("succeeds", func() {
				Expect(found).To(BeTrue())
				Expect(findErr).NotTo(HaveOccurred())
			})

			It("returns the created container", func() {
				Expect(foundContainer).To(Equal(fakeContainer))
			})
		})
	})

	Describe("FindVolume", func() {
		var (
			foundVolume Volume
			found       bool
			findErr     error
		)

		JustBeforeEach(func() {
			foundVolume, found, findErr = pool.FindVolume(
				logger,
				4567,
				"some-handle",
			)
		})

		Context("when looking up the worker errors", func() {
			BeforeEach(func() {
				fakeProvider.FindWorkerForVolumeReturns(nil, false, errors.New("nope"))
			})

			It("errors", func() {
				Expect(findErr).To(HaveOccurred())
			})
		})

		Context("when worker is not found", func() {
			BeforeEach(func() {
				fakeProvider.FindWorkerForVolumeReturns(nil, false, nil)
			})

			It("returns not found", func() {
				Expect(findErr).NotTo(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})

		Context("when a worker is found with the volume", func() {
			var fakeWorker *workerfakes.FakeWorker
			var fakeVolume *workerfakes.FakeVolume

			BeforeEach(func() {
				fakeWorker = new(workerfakes.FakeWorker)
				fakeProvider.FindWorkerForVolumeReturns(fakeWorker, true, nil)

				fakeVolume = new(workerfakes.FakeVolume)
				fakeWorker.LookupVolumeReturns(fakeVolume, true, nil)
			})

			It("succeeds", func() {
				Expect(found).To(BeTrue())
				Expect(findErr).NotTo(HaveOccurred())
			})

			It("returns the volume", func() {
				Expect(foundVolume).To(Equal(fakeVolume))
			})
		})
	})

	Describe("CreateVolume", func() {
		var (
			fakeWorker *workerfakes.FakeWorker
			volumeSpec VolumeSpec
			workerSpec WorkerSpec
			volumeType db.VolumeType
			err        error
		)

		BeforeEach(func() {
			volumeSpec = VolumeSpec{
				Strategy: baggageclaim.EmptyStrategy{},
			}

			workerSpec = WorkerSpec{
				TeamID: 1,
			}

			volumeType = db.VolumeTypeArtifact
		})

		JustBeforeEach(func() {
			_, err = pool.CreateVolume(logger, volumeSpec, workerSpec, volumeType)
		})

		Context("when no workers can be found", func() {
			BeforeEach(func() {
				fakeProvider.RunningWorkersReturns(nil, nil)
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the worker can be found", func() {
			BeforeEach(func() {
				fakeWorker = new(workerfakes.FakeWorker)
				fakeProvider.RunningWorkersReturns([]Worker{fakeWorker}, nil)
				fakeWorker.SatisfiesReturns(true)
			})

			It("creates the volume on the worker", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeWorker.CreateVolumeCallCount()).To(Equal(1))
				l, spec, id, t := fakeWorker.CreateVolumeArgsForCall(0)
				Expect(l).To(Equal(logger))
				Expect(spec).To(Equal(volumeSpec))
				Expect(id).To(Equal(1))
				Expect(t).To(Equal(volumeType))
			})
		})
	})

	Describe("SelectWorker", func() {
		var (
			spec       ContainerSpec
			workerSpec WorkerSpec
			fakeOwner  *dbfakes.FakeContainerOwner
			team       string
			pipeline   string

			chosenWorker Client
			chooseErr    error

			incompatibleWorker    *workerfakes.FakeWorker
			compatibleWorker      *workerfakes.FakeWorker
			otherCompatibleWorker *workerfakes.FakeWorker
			fakeStrategy          *workerfakes.FakeContainerPlacementStrategy
			fakeCallbacks         *workerfakes.FakePoolCallbacks
		)

		BeforeEach(func() {
			fakeStrategy = new(workerfakes.FakeContainerPlacementStrategy)
			fakeCallbacks = new(workerfakes.FakePoolCallbacks)

			fakeOwner = new(dbfakes.FakeContainerOwner)
			team = "some-team"
			pipeline = "some-pipeline"

			spec = ContainerSpec{
				ImageSpec: ImageSpec{ResourceType: "some-type"},
				TeamID:    4567,
			}

			workerSpec = WorkerSpec{
				ResourceType: "some-type",
				TeamID:       4567,
				Tags:         atc.Tags{"some-tag"},
			}

			incompatibleWorker = new(workerfakes.FakeWorker)
			incompatibleWorker.SatisfiesReturns(false)

			compatibleWorker = new(workerfakes.FakeWorker)
			compatibleWorker.SatisfiesReturns(true)

			otherCompatibleWorker = new(workerfakes.FakeWorker)
			otherCompatibleWorker.SatisfiesReturns(true)
		})

		JustBeforeEach(func() {
			chosenWorker, chooseErr = pool.SelectWorker(
				context.Background(),
				fakeOwner,
				team,
				pipeline,
				spec,
				workerSpec,
				fakeStrategy,
				fakeCallbacks,
			)
		})

		Context("when workers are found with the container", func() {
			var (
				workerA *workerfakes.FakeWorker
				workerB *workerfakes.FakeWorker
				workerC *workerfakes.FakeWorker
			)

			BeforeEach(func() {
				workerA = new(workerfakes.FakeWorker)
				workerA.NameReturns("workerA")
				workerB = new(workerfakes.FakeWorker)
				workerB.NameReturns("workerB")
				workerC = new(workerfakes.FakeWorker)
				workerC.NameReturns("workerC")

				fakeProvider.FindWorkersForContainerByOwnerReturns([]Worker{workerA, workerB, workerC}, nil)
				fakeProvider.RunningWorkersReturns([]Worker{workerA, workerB, workerC}, nil)
				fakeStrategy.CandidatesReturns([]Worker{workerA}, nil)
				fakeStrategy.PickReturns(nil)
			})

			Context("when a worker is found", func() {
				BeforeEach(func() {
					workerA.SatisfiesReturns(true)
				})

				It("calls the callback SelectWorker method", func() {
					Expect(fakeCallbacks.SelectedWorkerCallCount()).To(Equal(1))
					_, worker := fakeCallbacks.SelectedWorkerArgsForCall(0)
					Expect(worker).To(Equal(workerA))
				})
			})
			Context("when one of the workers satisfy the spec", func() {
				BeforeEach(func() {
					workerA.SatisfiesReturns(true)
					workerB.SatisfiesReturns(false)
					workerC.SatisfiesReturns(false)
				})

				It("succeeds and returns the compatible worker with the container", func() {
					Expect(fakeStrategy.CandidatesCallCount()).To(Equal(0))

					Expect(chooseErr).NotTo(HaveOccurred())
					Expect(chosenWorker.Name()).To(Equal(workerA.Name()))
				})
			})

			Context("when multiple workers satisfy the spec", func() {
				BeforeEach(func() {
					workerA.SatisfiesReturns(true)
					workerB.SatisfiesReturns(true)
					workerC.SatisfiesReturns(false)
				})

				It("succeeds and returns the first compatible worker with the container", func() {
					Expect(fakeStrategy.CandidatesCallCount()).To(Equal(0))

					Expect(chooseErr).NotTo(HaveOccurred())
					Expect(chosenWorker.Name()).To(Equal(workerA.Name()))
				})
			})

			Context("when the worker that has the container does not satisfy the spec", func() {
				BeforeEach(func() {
					workerA.SatisfiesReturns(true)
					workerB.SatisfiesReturns(true)
					workerC.SatisfiesReturns(false)

					fakeProvider.FindWorkersForContainerByOwnerReturns([]Worker{workerC}, nil)
				})

				It("chooses a satisfying worker", func() {
					Expect(fakeStrategy.CandidatesCallCount()).To(Equal(1))

					Expect(chooseErr).NotTo(HaveOccurred())
					Expect(chosenWorker.Name()).ToNot(Equal(workerC.Name()))
				})
			})
		})

		Context("when no worker is found with the container", func() {
			BeforeEach(func() {
				fakeProvider.FindWorkersForContainerByOwnerReturns(nil, nil)
			})

			Context("with multiple workers", func() {
				var (
					workerA *workerfakes.FakeWorker
					workerB *workerfakes.FakeWorker
					workerC *workerfakes.FakeWorker
				)

				BeforeEach(func() {
					workerA = new(workerfakes.FakeWorker)
					workerB = new(workerfakes.FakeWorker)
					workerC = new(workerfakes.FakeWorker)
					workerA.NameReturns("workerA")

					workerA.SatisfiesReturns(true)
					workerB.SatisfiesReturns(true)
					workerC.SatisfiesReturns(false)

					fakeProvider.RunningWorkersReturns([]Worker{workerA, workerB, workerC}, nil)
					fakeStrategy.CandidatesReturns([]Worker{workerA}, nil)
					fakeStrategy.PickReturns(nil)
				})

				It("checks that the workers satisfy the given worker spec", func() {
					Expect(workerA.SatisfiesCallCount()).To(Equal(1))
					_, actualSpec := workerA.SatisfiesArgsForCall(0)
					Expect(actualSpec).To(Equal(workerSpec))

					Expect(workerB.SatisfiesCallCount()).To(Equal(1))
					_, actualSpec = workerB.SatisfiesArgsForCall(0)
					Expect(actualSpec).To(Equal(workerSpec))

					Expect(workerC.SatisfiesCallCount()).To(Equal(1))
					_, actualSpec = workerC.SatisfiesArgsForCall(0)
					Expect(actualSpec).To(Equal(workerSpec))
				})

				It("returns all workers satisfying the spec", func() {
					_, satisfyingWorkers, _ := fakeStrategy.CandidatesArgsForCall(0)
					Expect(satisfyingWorkers).To(ConsistOf(workerA, workerB))
				})
			})

			Context("when team workers and general workers satisfy the spec", func() {
				var (
					teamWorker1   *workerfakes.FakeWorker
					teamWorker2   *workerfakes.FakeWorker
					teamWorker3   *workerfakes.FakeWorker
					generalWorker *workerfakes.FakeWorker
				)

				BeforeEach(func() {
					teamWorker1 = new(workerfakes.FakeWorker)
					teamWorker1.SatisfiesReturns(true)
					teamWorker1.IsOwnedByTeamReturns(true)
					teamWorker2 = new(workerfakes.FakeWorker)
					teamWorker2.SatisfiesReturns(true)
					teamWorker2.IsOwnedByTeamReturns(true)
					teamWorker3 = new(workerfakes.FakeWorker)
					teamWorker3.SatisfiesReturns(false)
					generalWorker = new(workerfakes.FakeWorker)
					generalWorker.SatisfiesReturns(true)
					generalWorker.IsOwnedByTeamReturns(false)
					fakeProvider.RunningWorkersReturns([]Worker{generalWorker, teamWorker1, teamWorker2, teamWorker3}, nil)
					fakeStrategy.CandidatesReturns([]Worker{teamWorker1}, nil)
					fakeStrategy.PickReturns(nil)
				})

				It("returns only the team workers that satisfy the spec", func() {
					_, satisfyingWorkers, _ := fakeStrategy.CandidatesArgsForCall(0)
					Expect(satisfyingWorkers).To(ConsistOf(teamWorker1, teamWorker2))
				})
			})

			Context("when only general workers satisfy the spec", func() {
				var (
					teamWorker     *workerfakes.FakeWorker
					generalWorker1 *workerfakes.FakeWorker
					generalWorker2 *workerfakes.FakeWorker
				)

				BeforeEach(func() {
					teamWorker = new(workerfakes.FakeWorker)
					teamWorker.SatisfiesReturns(false)
					generalWorker1 = new(workerfakes.FakeWorker)
					generalWorker1.SatisfiesReturns(true)
					generalWorker1.IsOwnedByTeamReturns(false)
					generalWorker2 = new(workerfakes.FakeWorker)
					generalWorker2.SatisfiesReturns(false)
					fakeProvider.RunningWorkersReturns([]Worker{generalWorker1, generalWorker2, teamWorker}, nil)
					fakeStrategy.CandidatesReturns([]Worker{generalWorker1}, nil)
					fakeStrategy.PickReturns(nil)
				})

				It("returns the general workers that satisfy the spec", func() {
					_, satisfyingWorkers, _ := fakeStrategy.CandidatesArgsForCall(0)
					Expect(satisfyingWorkers).To(ConsistOf(generalWorker1))
				})
			})

			Context("with no workers", func() {
				BeforeEach(func() {
					fakeProvider.RunningWorkersReturns([]Worker{}, nil)
				})

				It("returns ErrNoWorkers", func() {
					Expect(chooseErr).To(Equal(ErrNoWorkers))
				})
			})

			Context("when getting the workers fails", func() {
				disaster := errors.New("nope")

				BeforeEach(func() {
					fakeProvider.RunningWorkersReturns(nil, disaster)
				})

				It("returns the error", func() {
					Expect(chooseErr).To(Equal(disaster))
				})
			})

			Context("with no compatible workers available", func() {
				BeforeEach(func() {
					fakeProvider.RunningWorkersReturns([]Worker{incompatibleWorker}, nil)
				})

				It("returns NoCompatibleWorkersError", func() {
					Expect(chooseErr).To(Equal(NoCompatibleWorkersError{
						Spec: workerSpec,
					}))
				})
			})

			Context("with compatible workers available", func() {
				BeforeEach(func() {
					fakeProvider.RunningWorkersReturns([]Worker{
						incompatibleWorker,
						compatibleWorker,
						otherCompatibleWorker,
					}, nil)
				})

				Context("when strategy returns a worker", func() {
					BeforeEach(func() {
						fakeStrategy.CandidatesReturns([]Worker{compatibleWorker}, nil)
						fakeStrategy.PickReturns(nil)
					})

					It("chooses a worker", func() {
						Expect(chooseErr).ToNot(HaveOccurred())
						Expect(fakeStrategy.CandidatesCallCount()).To(Equal(1))
						Expect(chosenWorker.Name()).To(Equal(compatibleWorker.Name()))
					})
				})

				Context("when strategy returns multiple pickable workers", func() {
					BeforeEach(func() {
						fakeStrategy.CandidatesReturns([]Worker{compatibleWorker, otherCompatibleWorker}, nil)
						fakeStrategy.PickReturns(nil)
					})

					It("chooses the first worker", func() {
						Expect(chooseErr).ToNot(HaveOccurred())
						Expect(fakeStrategy.CandidatesCallCount()).To(Equal(1))
						Expect(chosenWorker.Name()).To(Equal(compatibleWorker.Name()))
					})
				})

				Context("when strategy candidates errors", func() {
					var (
						strategyError error
					)

					BeforeEach(func() {
						strategyError = errors.New("strategical explosion")
						fakeStrategy.CandidatesReturns(nil, strategyError)
					})

					It("returns an error", func() {
						Expect(chooseErr).To(Equal(strategyError))
					})
				})

				Context("when strategy returns some unpickable workers", func() {
					BeforeEach(func() {
						fakeStrategy.CandidatesReturns([]Worker{compatibleWorker, otherCompatibleWorker}, nil)
						fakeStrategy.PickReturnsOnCall(0, errors.New("can't pick this worker"))
						fakeStrategy.PickReturnsOnCall(1, nil)
					})

					It("chooses the first worker", func() {
						Expect(chooseErr).ToNot(HaveOccurred())
						Expect(fakeStrategy.CandidatesCallCount()).To(Equal(1))
						Expect(fakeStrategy.PickCallCount()).To(Equal(2))
						Expect(chosenWorker.Name()).To(Equal(otherCompatibleWorker.Name()))
					})
				})

				Context("when strategy returns no pickable workers", func() {
					BeforeEach(func() {
						fakeStrategy.CandidatesReturns([]Worker{compatibleWorker, otherCompatibleWorker}, nil)
						fakeStrategy.PickReturns(errors.New("can't pick this worker"))
					})

					It("returns an error", func() {
						Expect(chooseErr).To(Equal(ErrFailedToPickWorker))
					})
				})

				Context("when the policy checker is enabled", func() {
					BeforeEach(func() {
						fakeStrategy.CandidatesReturns([]Worker{compatibleWorker, otherCompatibleWorker}, nil)
						fakeStrategy.PickReturns(nil)
						fakePolicyChecker.ShouldCheckActionReturns(true)
					})

					Context("when all workers are valid", func() {
						BeforeEach(func() {
							fakePolicyChecker.CheckReturns(policy.PassedPolicyCheck(), nil)
						})

						It("selects the first worker", func() {
							Expect(chooseErr).ToNot(HaveOccurred())
							Expect(chosenWorker.Name()).To(Equal(compatibleWorker.Name()))
						})
					})

					Context("when a worker is not valid", func() {
						BeforeEach(func() {
							fakePolicyChecker.CheckReturnsOnCall(0, policy.FailedPolicyCheck(), nil)
							fakePolicyChecker.CheckReturnsOnCall(1, policy.PassedPolicyCheck(), nil)
						})

						It("selects the next worker", func() {
							Expect(chooseErr).ToNot(HaveOccurred())
							Expect(chosenWorker.Name()).To(Equal(otherCompatibleWorker.Name()))
						})
					})

					Context("when no workers are valid", func() {
						BeforeEach(func() {
							fakePolicyChecker.CheckReturns(policy.FailedPolicyCheck(), nil)
						})

						It("returns an error", func() {
							Expect(chooseErr).To(Equal(ErrFailedToPickWorker))
						})
					})

					Context("when the policy check errors", func() {
						BeforeEach(func() {
							fakePolicyChecker.CheckReturns(policy.FailedPolicyCheck(), errors.New("policy check failed"))
						})

						It("returns an error", func() {
							Expect(chooseErr).To(Equal(ErrFailedToPickWorker))
						})
					})
				})
			})
		})
	})
})
