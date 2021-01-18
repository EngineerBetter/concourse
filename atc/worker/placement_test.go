package worker_test

import (
	"errors"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	"github.com/concourse/concourse/atc"
	"github.com/concourse/concourse/atc/db"
	"github.com/concourse/concourse/atc/db/dbfakes"
	. "github.com/concourse/concourse/atc/worker"
	"github.com/concourse/concourse/atc/worker/workerfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//go:generate counterfeiter . ContainerPlacementStrategy

var (
	strategy ContainerPlacementStrategy

	spec     ContainerSpec
	metadata db.ContainerMetadata
	workers  []Worker

	chosenWorkers []Worker
	chooseErr     error

	newStrategyError error

	compatibleWorkerOneCache1 *workerfakes.FakeWorker
	compatibleWorkerOneCache2 *workerfakes.FakeWorker
	compatibleWorkerTwoCaches *workerfakes.FakeWorker
	compatibleWorkerNoCaches1 *workerfakes.FakeWorker
	compatibleWorkerNoCaches2 *workerfakes.FakeWorker

	logger *lagertest.TestLogger
)

var _ = Describe("FewestBuildContainersPlacementStrategy", func() {
	Describe("Candidates", func() {
		var compatibleWorker1 *workerfakes.FakeWorker
		var compatibleWorker2 *workerfakes.FakeWorker
		var compatibleWorker3 *workerfakes.FakeWorker

		BeforeEach(func() {
			logger = lagertest.NewTestLogger("build-containers-equal-placement-test")
			strategy, newStrategyError = NewContainerPlacementStrategy(ContainerPlacementStrategyOptions{ContainerPlacementStrategy: []string{"fewest-build-containers"}}, new(dbfakes.FakeContainerRepository))
			Expect(newStrategyError).ToNot(HaveOccurred())
			compatibleWorker1 = new(workerfakes.FakeWorker)
			compatibleWorker1.NameReturns("compatibleWorker1")
			compatibleWorker2 = new(workerfakes.FakeWorker)
			compatibleWorker2.NameReturns("compatibleWorker2")
			compatibleWorker3 = new(workerfakes.FakeWorker)
			compatibleWorker3.NameReturns("compatibleWorker3")

			spec = ContainerSpec{
				ImageSpec: ImageSpec{ResourceType: "some-type"},

				TeamID: 4567,

				Inputs: []InputSource{},
			}
		})

		Context("when there is only one worker", func() {
			BeforeEach(func() {
				workers = []Worker{compatibleWorker1}
				compatibleWorker1.BuildContainersReturns(20)
			})

			It("returns that worker", func() {
				chosenWorkers, chooseErr = strategy.Candidates(
					logger,
					workers,
					spec,
				)
				Expect(chooseErr).ToNot(HaveOccurred())
				Expect(chosenWorkers).To(ConsistOf(compatibleWorker1))
			})
		})

		Context("when there are multiple workers", func() {
			BeforeEach(func() {
				workers = []Worker{compatibleWorker1, compatibleWorker2, compatibleWorker3}

				compatibleWorker1.BuildContainersReturns(30)
				compatibleWorker2.BuildContainersReturns(20)
				compatibleWorker3.BuildContainersReturns(10)
			})

			Context("when the container is not of type 'check'", func() {
				It("returns the one with least amount of containers", func() {
					Consistently(func() []Worker {
						chosenWorkers, chooseErr = strategy.Candidates(
							logger,
							workers,
							spec,
						)
						Expect(chooseErr).ToNot(HaveOccurred())
						return chosenWorkers
					}).Should(ConsistOf(compatibleWorker3))
				})

				Context("when there is more than one worker with the same number of build containers", func() {
					BeforeEach(func() {
						workers = []Worker{compatibleWorker1, compatibleWorker2, compatibleWorker3}
						compatibleWorker1.BuildContainersReturns(10)
					})

					It("returns both of them", func() {
						Consistently(func() []Worker {
							chosenWorkers, chooseErr = strategy.Candidates(
								logger,
								workers,
								spec,
							)
							Expect(chooseErr).ToNot(HaveOccurred())
							return chosenWorkers
						}).Should(ConsistOf(compatibleWorker1, compatibleWorker3))
					})
				})

			})
		})
	})
})

var _ = Describe("VolumeLocalityPlacementStrategy", func() {
	Describe("Candidates", func() {
		JustBeforeEach(func() {
			chosenWorkers, chooseErr = strategy.Candidates(
				logger,
				workers,
				spec,
			)
		})

		BeforeEach(func() {
			logger = lagertest.NewTestLogger("volume-locality-placement-test")
			strategy, newStrategyError = NewContainerPlacementStrategy(ContainerPlacementStrategyOptions{ContainerPlacementStrategy: []string{"volume-locality"}}, new(dbfakes.FakeContainerRepository))
			Expect(newStrategyError).ToNot(HaveOccurred())

			fakeInput1 := new(workerfakes.FakeInputSource)
			fakeInput1AS := new(workerfakes.FakeArtifactSource)
			fakeInput1AS.ExistsOnStub = func(logger lager.Logger, worker Worker) (Volume, bool, error) {
				switch worker {
				case compatibleWorkerOneCache1, compatibleWorkerOneCache2, compatibleWorkerTwoCaches:
					return new(workerfakes.FakeVolume), true, nil
				default:
					return nil, false, nil
				}
			}
			fakeInput1.SourceReturns(fakeInput1AS)

			fakeInput2 := new(workerfakes.FakeInputSource)
			fakeInput2AS := new(workerfakes.FakeArtifactSource)
			fakeInput2AS.ExistsOnStub = func(logger lager.Logger, worker Worker) (Volume, bool, error) {
				switch worker {
				case compatibleWorkerTwoCaches:
					return new(workerfakes.FakeVolume), true, nil
				default:
					return nil, false, nil
				}
			}
			fakeInput2.SourceReturns(fakeInput2AS)

			spec = ContainerSpec{
				ImageSpec: ImageSpec{ResourceType: "some-type"},

				TeamID: 4567,

				Inputs: []InputSource{
					fakeInput1,
					fakeInput2,
				},
			}

			compatibleWorkerOneCache1 = new(workerfakes.FakeWorker)
			compatibleWorkerOneCache1.SatisfiesReturns(true)
			compatibleWorkerOneCache1.NameReturns("compatibleWorkerOneCache1")

			compatibleWorkerOneCache2 = new(workerfakes.FakeWorker)
			compatibleWorkerOneCache2.SatisfiesReturns(true)
			compatibleWorkerOneCache2.NameReturns("compatibleWorkerOneCache2")

			compatibleWorkerTwoCaches = new(workerfakes.FakeWorker)
			compatibleWorkerTwoCaches.SatisfiesReturns(true)
			compatibleWorkerTwoCaches.NameReturns("compatibleWorkerTwoCaches")

			compatibleWorkerNoCaches1 = new(workerfakes.FakeWorker)
			compatibleWorkerNoCaches1.SatisfiesReturns(true)
			compatibleWorkerNoCaches1.NameReturns("compatibleWorkerNoCaches1")

			compatibleWorkerNoCaches2 = new(workerfakes.FakeWorker)
			compatibleWorkerNoCaches2.SatisfiesReturns(true)
			compatibleWorkerNoCaches2.NameReturns("compatibleWorkerNoCaches2")
		})

		Context("with one having the most local caches", func() {
			BeforeEach(func() {
				workers = []Worker{
					compatibleWorkerOneCache1,
					compatibleWorkerTwoCaches,
					compatibleWorkerNoCaches1,
					compatibleWorkerNoCaches2,
				}
			})

			It("returns worker with the most caches", func() {
				Expect(chooseErr).ToNot(HaveOccurred())
				Expect(chosenWorkers).To(ConsistOf(compatibleWorkerTwoCaches))
			})
		})

		Context("with multiple with the same amount of local caches", func() {
			BeforeEach(func() {
				workers = []Worker{
					compatibleWorkerOneCache1,
					compatibleWorkerOneCache2,
					compatibleWorkerNoCaches1,
					compatibleWorkerNoCaches2,
				}
			})

			It("returns both workers with the most caches", func() {
				Expect(chooseErr).ToNot(HaveOccurred())
				Expect(chosenWorkers).To(ConsistOf(compatibleWorkerOneCache1, compatibleWorkerOneCache2))

				workerChoiceCounts := map[Worker]int{}

				for i := 0; i < 100; i++ {
					workers, err := strategy.Candidates(
						logger,
						workers,
						spec,
					)
					Expect(err).ToNot(HaveOccurred())
					Expect(chosenWorkers).To(ConsistOf(compatibleWorkerOneCache1, compatibleWorkerOneCache2))
					for _, worker := range workers {
						workerChoiceCounts[worker]++
					}
				}

				Expect(workerChoiceCounts[compatibleWorkerOneCache1]).ToNot(BeZero())
				Expect(workerChoiceCounts[compatibleWorkerOneCache2]).ToNot(BeZero())
				Expect(workerChoiceCounts[compatibleWorkerNoCaches1]).To(BeZero())
				Expect(workerChoiceCounts[compatibleWorkerNoCaches2]).To(BeZero())
			})
		})

		Context("with none having any local caches", func() {
			BeforeEach(func() {
				workers = []Worker{
					compatibleWorkerNoCaches1,
					compatibleWorkerNoCaches2,
				}
			})

			It("returns all of them", func() {
				Expect(chooseErr).ToNot(HaveOccurred())
				Expect(chosenWorkers).To(ConsistOf(compatibleWorkerNoCaches1, compatibleWorkerNoCaches2))

				workerChoiceCounts := map[Worker]int{}

				for i := 0; i < 100; i++ {
					workers, err := strategy.Candidates(
						logger,
						workers,
						spec,
					)
					Expect(err).ToNot(HaveOccurred())
					Expect(chosenWorkers).To(ConsistOf(compatibleWorkerNoCaches1, compatibleWorkerNoCaches2))
					for _, worker := range workers {
						workerChoiceCounts[worker]++
					}
				}

				Expect(workerChoiceCounts[compatibleWorkerNoCaches1]).ToNot(BeZero())
				Expect(workerChoiceCounts[compatibleWorkerNoCaches2]).ToNot(BeZero())
			})
		})
	})
})

var _ = Describe("No strategy should equal to random strategy", func() {
	Describe("Candidates", func() {
		JustBeforeEach(func() {
			chosenWorkers, chooseErr = strategy.Candidates(
				logger,
				workers,
				spec,
			)
		})

		BeforeEach(func() {
			strategy = NewRandomPlacementStrategy()

			workers = []Worker{
				compatibleWorkerNoCaches1,
				compatibleWorkerNoCaches2,
			}
		})

		It("creates it ona random one of them", func() {
			Expect(chooseErr).ToNot(HaveOccurred())
			Expect(chosenWorkers).To(ConsistOf(compatibleWorkerNoCaches1, compatibleWorkerNoCaches2))

			workerChoiceCounts := map[Worker]int{}

			for i := 0; i < 100; i++ {
				workers, err := strategy.Candidates(
					logger,
					workers,
					spec,
				)
				Expect(err).ToNot(HaveOccurred())
				Expect(chosenWorkers).To(ConsistOf(compatibleWorkerNoCaches1, compatibleWorkerNoCaches2))
				for _, worker := range workers {
					workerChoiceCounts[worker]++
				}
			}

			Expect(workerChoiceCounts[compatibleWorkerNoCaches1]).ToNot(BeZero())
			Expect(workerChoiceCounts[compatibleWorkerNoCaches2]).ToNot(BeZero())
		})
	})
})

// TODO: Fix XDescribe for LimitActiveTasksPlacementStrategy
var _ = XDescribe("LimitActiveTasksPlacementStrategy", func() {
	Describe("Candidates", func() {
		var compatibleWorker1 *workerfakes.FakeWorker
		var compatibleWorker2 *workerfakes.FakeWorker
		var compatibleWorker3 *workerfakes.FakeWorker

		Context("when MaxActiveTasksPerWorker less than 0", func() {
			BeforeEach(func() {
				logger = lagertest.NewTestLogger("active-tasks-equal-placement-test")
				strategy, newStrategyError = NewContainerPlacementStrategy(ContainerPlacementStrategyOptions{ContainerPlacementStrategy: []string{"limit-active-tasks"}, MaxActiveTasksPerWorker: -1}, new(dbfakes.FakeContainerRepository))
			})

			It("should fail", func() {
				Expect(newStrategyError).To(HaveOccurred())
				Expect(newStrategyError).To(Equal(errors.New("max-active-tasks-per-worker must be greater or equal than 0")))
				Expect(strategy).To(BeNil())
			})
		})

		Context("when MaxActiveTasksPerWorker is 0", func() {
			BeforeEach(func() {
				logger = lagertest.NewTestLogger("active-tasks-equal-placement-test")
				strategy, newStrategyError = NewContainerPlacementStrategy(ContainerPlacementStrategyOptions{ContainerPlacementStrategy: []string{"limit-active-tasks"}, MaxActiveTasksPerWorker: 0}, new(dbfakes.FakeContainerRepository))
				Expect(newStrategyError).ToNot(HaveOccurred())

				compatibleWorker1 = new(workerfakes.FakeWorker)
				compatibleWorker1.NameReturns("compatibleWorker1")
				compatibleWorker2 = new(workerfakes.FakeWorker)
				compatibleWorker2.NameReturns("compatibleWorker2")
				compatibleWorker3 = new(workerfakes.FakeWorker)
				compatibleWorker3.NameReturns("compatibleWorker3")

				spec = ContainerSpec{
					ImageSpec: ImageSpec{ResourceType: "some-type"},
					Type:      "task",
					TeamID:    4567,
					Inputs:    []InputSource{},
				}
			})

			Context("when there is only one worker with any amount of running tasks", func() {
				BeforeEach(func() {
					workers = []Worker{compatibleWorker1}
					compatibleWorker1.ActiveTasksReturns(42, nil)
				})

				It("picks that worker", func() {
					chosenWorkers, chooseErr = strategy.Candidates(
						logger,
						workers,
						spec,
					)
					Expect(chooseErr).ToNot(HaveOccurred())
					Expect(chosenWorkers).To(ConsistOf(compatibleWorker1))
				})
			})

			Context("when there are multiple workers", func() {
				BeforeEach(func() {
					workers = []Worker{compatibleWorker1, compatibleWorker2, compatibleWorker3}

					compatibleWorker1.ActiveTasksReturns(2, nil)
					compatibleWorker2.ActiveTasksReturns(1, nil)
					compatibleWorker3.ActiveTasksReturns(2, nil)
				})

				It("a task picks the one with least active tasks", func() {
					Consistently(func() []Worker {
						chosenWorkers, chooseErr = strategy.Candidates(
							logger,
							workers,
							spec,
						)
						Expect(chooseErr).ToNot(HaveOccurred())
						return chosenWorkers
					}).Should(Equal(compatibleWorker2))
				})

				Context("when all the workers have the same number of active tasks", func() {
					BeforeEach(func() {
						workers = []Worker{compatibleWorker1, compatibleWorker2, compatibleWorker3}
						compatibleWorker1.ActiveTasksReturns(1, nil)
						compatibleWorker2.ActiveTasksReturns(1, nil)
						compatibleWorker3.ActiveTasksReturns(1, nil)
					})

					It("it returns all of them", func() {
						Consistently(func() []Worker {
							chosenWorkers, chooseErr = strategy.Candidates(
								logger,
								workers,
								spec,
							)
							Expect(chooseErr).ToNot(HaveOccurred())
							return chosenWorkers
						}).Should(ConsistOf(compatibleWorker1, compatibleWorker2, compatibleWorker3))
					})
				})
			})

			Context("when max-tasks-per-worker is set to 1", func() {
				BeforeEach(func() {
					strategy, newStrategyError = NewContainerPlacementStrategy(ContainerPlacementStrategyOptions{ContainerPlacementStrategy: []string{"limit-active-tasks"}, MaxActiveTasksPerWorker: 1}, new(dbfakes.FakeContainerRepository))
					Expect(newStrategyError).ToNot(HaveOccurred())
				})

				Context("when there are multiple workers", func() {
					BeforeEach(func() {
						workers = []Worker{compatibleWorker1, compatibleWorker2, compatibleWorker3}

						compatibleWorker1.ActiveTasksReturns(1, nil)
						compatibleWorker2.ActiveTasksReturns(0, nil)
						compatibleWorker3.ActiveTasksReturns(1, nil)
					})

					It("picks the worker with no active tasks", func() {
						chosenWorkers, chooseErr = strategy.Candidates(
							logger,
							workers,
							spec,
						)
						Expect(chooseErr).ToNot(HaveOccurred())
						Expect(chosenWorkers).To(ConsistOf(compatibleWorker2))
					})
				})

				Context("when all workers have active tasks", func() {
					BeforeEach(func() {
						workers = []Worker{compatibleWorker1, compatibleWorker2, compatibleWorker3}

						compatibleWorker1.ActiveTasksReturns(1, nil)
						compatibleWorker2.ActiveTasksReturns(1, nil)
						compatibleWorker3.ActiveTasksReturns(1, nil)
					})

					It("picks no worker", func() {
						chosenWorkers, chooseErr = strategy.Candidates(
							logger,
							workers,
							spec,
						)
						Expect(chooseErr).To(HaveOccurred())
						Expect(chooseErr).To(Equal(NoWorkerFitContainerPlacementStrategyError{Strategy: "limit-active-tasks"}))
						Expect(chosenWorkers).To(BeEmpty())
					})

					Context("when the container is not of type 'task'", func() {
						BeforeEach(func() {
							spec.Type = ""
						})

						It("picks any worker", func() {
							Consistently(func() []Worker {
								chosenWorkers, chooseErr = strategy.Candidates(
									logger,
									workers,
									spec,
								)
								Expect(chooseErr).ToNot(HaveOccurred())
								return chosenWorkers
							}).Should(ConsistOf(compatibleWorker1, compatibleWorker2, compatibleWorker3))
						})
					})
				})
			})
		})
	})
})

var _ = Describe("LimitActiveContainersPlacementStrategyNode", func() {
	Describe("Candidates", func() {
		var compatibleWorker1 *workerfakes.FakeWorker
		var compatibleWorker2 *workerfakes.FakeWorker
		var compatibleWorker3 *workerfakes.FakeWorker
		var activeContainerLimit int

		BeforeEach(func() {
			logger = lagertest.NewTestLogger("build-containers-equal-placement-test")
			compatibleWorker1 = new(workerfakes.FakeWorker)
			compatibleWorker1.NameReturns("compatibleWorker1")
			compatibleWorker2 = new(workerfakes.FakeWorker)
			compatibleWorker2.NameReturns("compatibleWorker2")
			compatibleWorker3 = new(workerfakes.FakeWorker)
			compatibleWorker3.NameReturns("compatibleWorker3")
			activeContainerLimit = 0

			compatibleWorker1.ActiveContainersReturns(20)
			compatibleWorker2.ActiveContainersReturns(200)
			compatibleWorker3.ActiveContainersReturns(200000000000)
			workers = []Worker{compatibleWorker1, compatibleWorker2, compatibleWorker3}

			spec = ContainerSpec{
				ImageSpec: ImageSpec{ResourceType: "some-type"},
				TeamID:    4567,
				Inputs:    []InputSource{},
			}
		})

		JustBeforeEach(func() {
			strategy, newStrategyError = NewContainerPlacementStrategy(ContainerPlacementStrategyOptions{
				ContainerPlacementStrategy:   []string{"limit-active-containers"},
				MaxActiveContainersPerWorker: activeContainerLimit,
			}, new(dbfakes.FakeContainerRepository),
			)
			Expect(newStrategyError).ToNot(HaveOccurred())
		})

		Context("when there is no limit", func() {
			BeforeEach(func() {
				activeContainerLimit = 0
			})

			It("return all workers", func() {
				Consistently(func() []Worker {
					chosenWorkers, chooseErr = strategy.Candidates(
						logger,
						workers,
						spec,
					)
					Expect(chooseErr).ToNot(HaveOccurred())
					return chosenWorkers
				}).Should(ConsistOf(compatibleWorker1, compatibleWorker2, compatibleWorker3))
			})
		})

		Context("when there is a limit", func() {
			Context("when the limit is 20", func() {
				BeforeEach(func() {
					activeContainerLimit = 20
				})

				It("picks worker1", func() {
					Consistently(func() []Worker {
						chosenWorkers, chooseErr = strategy.Candidates(
							logger,
							workers,
							spec,
						)
						Expect(chooseErr).ToNot(HaveOccurred())
						return chosenWorkers
					}).Should(ConsistOf(compatibleWorker1))
				})

				Context("when the limit is 200", func() {
					BeforeEach(func() {
						activeContainerLimit = 200
					})

					It("picks worker1 or worker2", func() {
						Consistently(func() []Worker {
							chosenWorkers, chooseErr = strategy.Candidates(
								logger,
								workers,
								spec,
							)
							Expect(chooseErr).ToNot(HaveOccurred())
							return chosenWorkers
						}).Should(ConsistOf(compatibleWorker1, compatibleWorker2))
					})
				})

				Context("when the limit is too low", func() {
					BeforeEach(func() {
						activeContainerLimit = 1
					})

					It("return no worker", func() {
						chosenWorkers, chooseErr = strategy.Candidates(
							logger,
							workers,
							spec,
						)
						Expect(chooseErr).To(HaveOccurred())
						Expect(chooseErr).To(Equal(NoWorkerFitContainerPlacementStrategyError{Strategy: "limit-active-containers"}))
						Expect(chosenWorkers).To(BeEmpty())
					})
				})
			})
		})
	})
})

var _ = Describe("LimitActiveVolumesPlacementStrategyNode", func() {
	Describe("Candidates", func() {
		var compatibleWorker1 *workerfakes.FakeWorker
		var compatibleWorker2 *workerfakes.FakeWorker
		var compatibleWorker3 *workerfakes.FakeWorker
		var activeVolumeLimit int

		BeforeEach(func() {
			logger = lagertest.NewTestLogger("build-containers-equal-placement-test")
			compatibleWorker1 = new(workerfakes.FakeWorker)
			compatibleWorker1.NameReturns("compatibleWorker1")
			compatibleWorker2 = new(workerfakes.FakeWorker)
			compatibleWorker2.NameReturns("compatibleWorker2")
			compatibleWorker3 = new(workerfakes.FakeWorker)
			compatibleWorker3.NameReturns("compatibleWorker3")
			activeVolumeLimit = 0

			compatibleWorker1.ActiveVolumesReturns(20)
			compatibleWorker2.ActiveVolumesReturns(200)
			compatibleWorker3.ActiveVolumesReturns(200000000000)
			workers = []Worker{compatibleWorker1, compatibleWorker2, compatibleWorker3}

			spec = ContainerSpec{
				ImageSpec: ImageSpec{ResourceType: "some-type"},
				TeamID:    4567,
				Inputs:    []InputSource{},
			}
		})

		JustBeforeEach(func() {
			strategy, newStrategyError = NewContainerPlacementStrategy(ContainerPlacementStrategyOptions{
				ContainerPlacementStrategy: []string{"limit-active-volumes"},
				MaxActiveVolumesPerWorker:  activeVolumeLimit,
			}, new(dbfakes.FakeContainerRepository),
			)
			Expect(newStrategyError).ToNot(HaveOccurred())
		})

		Context("when there is no limit", func() {
			BeforeEach(func() {
				activeVolumeLimit = 0
			})

			It("return all workers", func() {
				Consistently(func() []Worker {
					chosenWorkers, chooseErr = strategy.Candidates(
						logger,
						workers,
						spec,
					)
					Expect(chooseErr).ToNot(HaveOccurred())
					return chosenWorkers
				}).Should(ConsistOf(compatibleWorker1, compatibleWorker2, compatibleWorker3))
			})
		})
		Context("when there is a limit", func() {
			Context("when the limit is 20", func() {
				BeforeEach(func() {
					activeVolumeLimit = 20
				})

				It("picks worker1", func() {
					Consistently(func() []Worker {
						chosenWorkers, chooseErr = strategy.Candidates(
							logger,
							workers,
							spec,
						)
						Expect(chooseErr).ToNot(HaveOccurred())
						return chosenWorkers
					}).Should(ConsistOf(compatibleWorker1))
				})

				Context("when the limit is 200", func() {
					BeforeEach(func() {
						activeVolumeLimit = 200
					})

					It("picks worker1 or worker2", func() {
						Consistently(func() []Worker {
							chosenWorkers, chooseErr = strategy.Candidates(
								logger,
								workers,
								spec,
							)
							Expect(chooseErr).ToNot(HaveOccurred())
							return chosenWorkers
						}).Should(ConsistOf(compatibleWorker1, compatibleWorker2))
					})
				})

				Context("when the limit is too low", func() {
					BeforeEach(func() {
						activeVolumeLimit = 1
					})

					It("return no worker", func() {
						chosenWorkers, chooseErr = strategy.Candidates(
							logger,
							workers,
							spec,
						)
						Expect(chooseErr).To(HaveOccurred())
						Expect(chooseErr).To(Equal(NoWorkerFitContainerPlacementStrategyError{Strategy: "limit-active-volumes"}))
						Expect(chosenWorkers).To(BeEmpty())
					})
				})
			})
		})
	})
})

var _ = Describe("LimitTotalAllocatedMemoryPlacementStrategy", func() {
	Describe("Candidates", func() {
		var compatibleWorker1 *workerfakes.FakeWorker
		var compatibleWorker2 *workerfakes.FakeWorker
		var worker1AllocatedMemory int
		var worker2AllocatedMemory int
		var requestedMemory uint64
		var containerRepository *dbfakes.FakeContainerRepository

		BeforeEach(func() {
			logger = lagertest.NewTestLogger("build-containers-total-allocated-memory-test")

			compatibleWorker1 = new(workerfakes.FakeWorker)
			compatibleWorker1.NameReturns("compatibleWorker1")
			worker1AllocatableMemory := atc.MemoryLimit(2000)
			compatibleWorker1.AllocatableMemoryReturns(&worker1AllocatableMemory)
			compatibleWorker2 = new(workerfakes.FakeWorker)
			compatibleWorker2.NameReturns("compatibleWorker2")
			worker2AllocatableMemory := atc.MemoryLimit(2000)
			compatibleWorker2.AllocatableMemoryReturns(&worker2AllocatableMemory)

			containerRepository = new(dbfakes.FakeContainerRepository)
			containerRepository.GetActiveContainerMemoryAllocationStub = func(worker string) (atc.MemoryLimit, error) {
				switch worker {
				case "compatibleWorker1":
					return atc.MemoryLimit(worker1AllocatedMemory), nil
				case "compatibleWorker2":
					return atc.MemoryLimit(worker2AllocatedMemory), nil
				default:
					return 0, nil
				}
			}
			workers = []Worker{compatibleWorker1, compatibleWorker2}
		})

		JustBeforeEach(func() {
			strategy, newStrategyError = NewContainerPlacementStrategy(ContainerPlacementStrategyOptions{
				ContainerPlacementStrategy: []string{"limit-total-allocated-memory"},
			}, containerRepository)
			Expect(newStrategyError).ToNot(HaveOccurred())

			spec = ContainerSpec{
				ImageSpec: ImageSpec{ResourceType: "some-type"},
				TeamID:    4567,
				Inputs:    []InputSource{},
				Limits: ContainerLimits{
					Memory: &requestedMemory,
				},
			}
		})

		Context("when the memory is below the max limit", func() {
			BeforeEach(func() {
				worker1AllocatedMemory = 0
				worker2AllocatedMemory = 0
				requestedMemory = 1500
			})

			It("return all workers", func() {
				Consistently(func() []Worker {
					chosenWorkers, chooseErr = strategy.Candidates(
						logger,
						workers,
						spec,
					)
					Expect(chooseErr).ToNot(HaveOccurred())
					return chosenWorkers
				}).Should(ConsistOf(compatibleWorker1, compatibleWorker2))
			})
		})

		Context("when the memory is above the max limit", func() {
			BeforeEach(func() {
				worker1AllocatedMemory = 1500
				worker2AllocatedMemory = 0
				requestedMemory = 1000
			})

			It("return only workers with enough available memory", func() {
				Consistently(func() []Worker {
					chosenWorkers, chooseErr = strategy.Candidates(
						logger,
						workers,
						spec,
					)
					Expect(chooseErr).ToNot(HaveOccurred())
					return chosenWorkers
				}).Should(ConsistOf(compatibleWorker2))
			})
		})

		Context("when a worker does not configure allocatable memory", func() {
			var noAllocatableMemoryWorker *workerfakes.FakeWorker

			BeforeEach(func() {
				worker1AllocatedMemory = 1500
				worker2AllocatedMemory = 1500
				requestedMemory = 1000
				noAllocatableMemoryWorker = new(workerfakes.FakeWorker)
				noAllocatableMemoryWorker.NameReturns("noAllocatableMemoryWorker")
				noAllocatableMemoryWorker.AllocatableMemoryReturns(nil)
			})

			It("always returns workers without a configured allocatable memory", func() {
				Consistently(func() []Worker {
					chosenWorkers, chooseErr = strategy.Candidates(
						logger,
						[]Worker{compatibleWorker1, compatibleWorker2, noAllocatableMemoryWorker},
						spec,
					)
					Expect(chooseErr).ToNot(HaveOccurred())
					return chosenWorkers
				}).Should(ConsistOf(noAllocatableMemoryWorker))
			})

			It("does not get allocated resources for workers without a configured allocatable memory", func() {
				chosenWorkers, chooseErr = strategy.Candidates(
					logger,
					[]Worker{noAllocatableMemoryWorker},
					spec,
				)
				Expect(chooseErr).ToNot(HaveOccurred())
				Expect(containerRepository.GetActiveContainerMemoryAllocationCallCount()).To(Equal(0))
			})
		})
	})
})

var _ = Describe("ChainedPlacementStrategy #Candidates", func() {

	var someWorker1 *workerfakes.FakeWorker
	var someWorker2 *workerfakes.FakeWorker
	var someWorker3 *workerfakes.FakeWorker

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("build-containers-equal-placement-test")
		strategy, newStrategyError = NewContainerPlacementStrategy(
			ContainerPlacementStrategyOptions{
				ContainerPlacementStrategy: []string{"fewest-build-containers", "volume-locality"},
			}, new(dbfakes.FakeContainerRepository))
		Expect(newStrategyError).ToNot(HaveOccurred())
		someWorker1 = new(workerfakes.FakeWorker)
		someWorker1.NameReturns("worker1")
		someWorker2 = new(workerfakes.FakeWorker)
		someWorker2.NameReturns("worker2")
		someWorker3 = new(workerfakes.FakeWorker)
		someWorker3.NameReturns("worker3")

		spec = ContainerSpec{
			ImageSpec: ImageSpec{ResourceType: "some-type"},

			TeamID: 4567,

			Inputs: []InputSource{},
		}
	})

	Context("when there are multiple workers", func() {
		BeforeEach(func() {
			workers = []Worker{someWorker1, someWorker2, someWorker3}

			someWorker1.BuildContainersReturns(30)
			someWorker2.BuildContainersReturns(20)
			someWorker3.BuildContainersReturns(10)
		})

		It("picks the one with least amount of containers", func() {
			Consistently(func() []Worker {
				chosenWorkers, chooseErr = strategy.Candidates(
					logger,
					workers,
					spec,
				)
				Expect(chooseErr).ToNot(HaveOccurred())
				return chosenWorkers
			}).Should(ConsistOf(someWorker3))
		})

		Context("when there is more than one worker with the same number of build containers", func() {
			BeforeEach(func() {
				workers = []Worker{someWorker1, someWorker2, someWorker3}
				someWorker1.BuildContainersReturns(10)
				someWorker2.BuildContainersReturns(20)
				someWorker3.BuildContainersReturns(10)

				fakeInput1 := new(workerfakes.FakeInputSource)
				fakeInput1AS := new(workerfakes.FakeArtifactSource)
				fakeInput1AS.ExistsOnStub = func(logger lager.Logger, worker Worker) (Volume, bool, error) {
					switch worker {
					case someWorker3:
						return new(workerfakes.FakeVolume), true, nil
					default:
						return nil, false, nil
					}
				}
				fakeInput1.SourceReturns(fakeInput1AS)

				spec = ContainerSpec{
					ImageSpec: ImageSpec{ResourceType: "some-type"},

					TeamID: 4567,

					Inputs: []InputSource{
						fakeInput1,
					},
				}
			})
			It("picks the one with the most volumes", func() {
				Consistently(func() []Worker {
					cWorkers, cErr := strategy.Candidates(
						logger,
						workers,
						spec,
					)
					Expect(cErr).ToNot(HaveOccurred())
					return cWorkers
				}).Should(ConsistOf(someWorker3))

			})
		})

	})
})
