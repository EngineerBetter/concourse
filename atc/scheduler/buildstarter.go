package scheduler

import (
	"context"
	"fmt"

	"code.cloudfoundry.org/lager"
	"github.com/concourse/concourse/atc"
	"github.com/concourse/concourse/atc/db"
	"github.com/concourse/concourse/atc/metric"
	"github.com/concourse/concourse/atc/scheduler/algorithm"
)

//go:generate counterfeiter . BuildStarter

type BuildStarter interface {
	TryStartPendingBuildsForJob(
		logger lager.Logger,
		pipeline db.Pipeline,
		job db.Job,
		inputs []atc.JobInput,
		resources db.Resources,
		relatedJobs algorithm.NameToIDMap,
	) (bool, error)
}

//go:generate counterfeiter . BuildFactory

type BuildFactory interface {
	Create(atc.JobConfig, atc.ResourceConfigs, atc.VersionedResourceTypes, []db.BuildInput) (atc.Plan, error)
}

type Build interface {
	db.Build

	IsReadyToDetermineInputs(lager.Logger) bool
	BuildInputs(context.Context) ([]db.BuildInput, bool, error)
}

func NewBuildStarter(
	factory BuildFactory,
	algorithm Algorithm,
	eventProcessor db.EventProcessor,
) BuildStarter {
	return &buildStarter{
		factory:        factory,
		algorithm:      algorithm,
		eventProcessor: eventProcessor,
	}
}

type buildStarter struct {
	factory        BuildFactory
	algorithm      Algorithm
	eventProcessor db.EventProcessor
}

func (s *buildStarter) TryStartPendingBuildsForJob(
	logger lager.Logger,
	pipeline db.Pipeline,
	job db.Job,
	jobInputs []atc.JobInput,
	resources db.Resources,
	relatedJobs algorithm.NameToIDMap,
) (bool, error) {
	nextPendingBuilds, err := job.GetPendingBuilds()
	if err != nil {
		return false, fmt.Errorf("get pending builds: %w", err)
	}

	buildsToSchedule := s.constructBuilds(job, jobInputs, resources, relatedJobs, nextPendingBuilds)

	var needsRetry bool
	for _, nextSchedulableBuild := range buildsToSchedule {
		results, err := s.tryStartNextPendingBuild(logger, pipeline, nextSchedulableBuild, job, resources)
		if err != nil {
			return false, err
		}

		if results.finished {
			// If the build is successfully aborted, errored or started, continue
			// onto the next pending build
			continue
		}

		if !results.scheduled || !results.readyToDetermineInputs {
			// If max in flight is reached or a manually triggered build has not
			// checked all resources, stop scheduling and retry later
			needsRetry = true
			break
		}

		if !results.inputsDetermined {
			if nextSchedulableBuild.RerunOf() != 0 {
				// If it is a rerun build, continue on to next build. We don't want to
				// stop scheduling other builds because of a rerun build cannot
				// determine inputs
				continue
			} else {
				// If it is a regular scheduler build, stop scheduling because it is
				// failing to determine inputs
				break
			}
		}
	}

	return needsRetry, nil
}

func (s *buildStarter) constructBuilds(job db.Job, jobInputs []atc.JobInput, resources db.Resources, relatedJobIDs map[string]int, builds []db.Build) []Build {
	var buildsToSchedule []Build

	for _, nextPendingBuild := range builds {
		if nextPendingBuild.IsManuallyTriggered() {
			buildsToSchedule = append(buildsToSchedule, &manualTriggerBuild{
				Build:         nextPendingBuild,
				algorithm:     s.algorithm,
				job:           job,
				jobInputs:     jobInputs,
				resources:     resources,
				relatedJobIDs: relatedJobIDs,
			})
		} else if nextPendingBuild.RerunOf() != 0 {
			buildsToSchedule = append(buildsToSchedule, &rerunBuild{
				Build: nextPendingBuild,
			})
		} else {
			buildsToSchedule = append(buildsToSchedule, &schedulerBuild{
				Build: nextPendingBuild,
			})
		}
	}

	return buildsToSchedule
}

type startResults struct {
	finished               bool
	scheduled              bool
	readyToDetermineInputs bool
	inputsDetermined       bool
}

func (s *buildStarter) tryStartNextPendingBuild(
	logger lager.Logger,
	pipeline db.Pipeline,
	nextPendingBuild Build,
	job db.Job,
	resources db.Resources,
) (startResults, error) {
	logger = logger.Session("try-start-next-pending-build", lager.Data{
		"build-id":   nextPendingBuild.ID(),
		"build-name": nextPendingBuild.Name(),
	})

	if nextPendingBuild.IsAborted() {
		logger.Debug("cancel-aborted-pending-build")

		err := nextPendingBuild.Finish(db.BuildStatusAborted, s.eventProcessor)
		if err != nil {
			return startResults{}, fmt.Errorf("finish aborted build: %w", err)
		}

		return startResults{
			finished: true,
		}, nil
	}

	pipelinePaused, err := pipeline.CheckPaused()
	if err != nil {
		return startResults{}, fmt.Errorf("check pipeline paused: %w", err)
	}

	if pipelinePaused {
		logger.Debug("pipeline-paused")
		return startResults{}, nil
	}

	if job.Paused() {
		logger.Debug("job-paused")
		return startResults{}, nil
	}

	scheduled, err := job.ScheduleBuild(nextPendingBuild)
	if err != nil {
		return startResults{}, fmt.Errorf("schedule build: %w", err)
	}

	if !scheduled {
		logger.Debug("build-not-scheduled")
		return startResults{
			scheduled: scheduled,
		}, nil
	}

	readyToDetermineInputs := nextPendingBuild.IsReadyToDetermineInputs(logger)
	if !readyToDetermineInputs {
		return startResults{
			scheduled:              scheduled,
			readyToDetermineInputs: readyToDetermineInputs,
		}, nil
	}

	buildInputs, inputsDetermined, err := nextPendingBuild.BuildInputs(context.TODO())
	if err != nil {
		return startResults{}, fmt.Errorf("get build inputs: %w", err)
	}

	if !inputsDetermined {
		logger.Debug("build-inputs-not-found")

		// don't retry when build inputs are not found because this is due to the
		// inputs being unsatisfiable
		return startResults{
			scheduled:              scheduled,
			readyToDetermineInputs: readyToDetermineInputs,
			inputsDetermined:       inputsDetermined,
		}, nil
	}

	resourceTypes, err := pipeline.ResourceTypes()
	if err != nil {
		return startResults{}, fmt.Errorf("find resource types: %w", err)
	}

	resourceConfigs := atc.ResourceConfigs{}
	for _, v := range resources {
		resourceConfigs = append(resourceConfigs, atc.ResourceConfig{
			Name:   v.Name(),
			Type:   v.Type(),
			Source: v.Source(),
			Tags:   v.Tags(),
		})
	}

	config, err := job.Config()
	if err != nil {
		return startResults{}, fmt.Errorf("config: %w", err)
	}

	plan, err := s.factory.Create(config, resourceConfigs, resourceTypes.Deserialize(), buildInputs)
	if err != nil {
		logger.Error("failed-to-create-build-plan", err)

		// Don't use ErrorBuild because it logs a build event, and this build hasn't started
		if err = nextPendingBuild.Finish(db.BuildStatusErrored, s.eventProcessor); err != nil {
			logger.Error("failed-to-mark-build-as-errored", err)
			return startResults{}, fmt.Errorf("finish build: %w", err)
		}

		return startResults{
			finished: true,
		}, nil
	}

	started, err := nextPendingBuild.Start(plan, s.eventProcessor)
	if err != nil {
		logger.Error("failed-to-mark-build-as-started", err)
		return startResults{}, fmt.Errorf("start build: %w", err)
	}

	if !started {
		if err = nextPendingBuild.Finish(db.BuildStatusAborted, s.eventProcessor); err != nil {
			logger.Error("failed-to-mark-build-as-finished", err)
			return startResults{}, fmt.Errorf("finish build: %w", err)
		}

		return startResults{
			finished: true,
		}, nil
	}

	metric.BuildsStarted.Inc()

	return startResults{
		finished: true,
	}, nil
}
