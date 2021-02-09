package worker

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagerctx"

	"github.com/concourse/concourse/atc/db"
	"github.com/concourse/concourse/atc/policy"
)

var (
	ErrNoWorkers             = errors.New("no workers")
	ErrFailedAcquirePoolLock = errors.New("failed to acquire pool lock")
)

type NoCompatibleWorkersError struct {
	Spec WorkerSpec
}

func (err NoCompatibleWorkersError) Error() string {
	return fmt.Sprintf("no workers satisfying: %s", err.Spec.Description())
}

type NoValidWorkersError struct {
	Reasons []string
}

func (err NoValidWorkersError) Error() string {
	return fmt.Sprintf("no workers passed policy check: %s", err.Reasons)
}

//go:generate counterfeiter . Pool

type Pool interface {
	FindContainer(lager.Logger, int, string) (Container, bool, error)
	VolumeFinder
	CreateVolume(lager.Logger, VolumeSpec, WorkerSpec, db.VolumeType) (Volume, error)

	ContainerInWorker(lager.Logger, db.ContainerOwner, WorkerSpec) (bool, error)

	SelectWorker(
		context.Context,
		db.ContainerOwner,
		ContainerSpec,
		WorkerSpec,
		ContainerPlacementStrategy,
	) (Client, error)
}

//go:generate counterfeiter . VolumeFinder

type VolumeFinder interface {
	FindVolume(lager.Logger, int, string) (Volume, bool, error)
}

type pool struct {
	provider      WorkerProvider
	rand          *rand.Rand
	policyChecker policy.Checker
}

func NewPool(
	provider WorkerProvider,
	policyChecker policy.Checker,
) Pool {
	return &pool{
		provider:      provider,
		rand:          rand.New(rand.NewSource(time.Now().UnixNano())),
		policyChecker: policyChecker,
	}
}

func (pool *pool) allSatisfying(logger lager.Logger, spec WorkerSpec) ([]Worker, error) {
	workers, err := pool.provider.RunningWorkers(logger)
	if err != nil {
		return nil, err
	}

	if len(workers) == 0 {
		return nil, ErrNoWorkers
	}

	compatibleTeamWorkers := []Worker{}
	compatibleGeneralWorkers := []Worker{}
	for _, worker := range workers {
		compatible := worker.Satisfies(logger, spec)
		if compatible {
			if worker.IsOwnedByTeam() {
				compatibleTeamWorkers = append(compatibleTeamWorkers, worker)
			} else {
				compatibleGeneralWorkers = append(compatibleGeneralWorkers, worker)
			}
		}
	}

	if len(compatibleTeamWorkers) != 0 {
		// XXX(aoldershaw): if there is a team worker that is compatible but is
		// rejected by the strategy, shouldn't we fallback to general workers?
		return compatibleTeamWorkers, nil
	}

	if len(compatibleGeneralWorkers) != 0 {
		return compatibleGeneralWorkers, nil
	}

	return nil, NoCompatibleWorkersError{
		Spec: spec,
	}
}

func (pool *pool) FindContainer(logger lager.Logger, teamID int, handle string) (Container, bool, error) {
	worker, found, err := pool.provider.FindWorkerForContainer(
		logger.Session("find-worker"),
		teamID,
		handle,
	)
	if err != nil {
		return nil, false, err
	}

	if !found {
		return nil, false, nil
	}

	return worker.FindContainerByHandle(logger, teamID, handle)
}

func (pool *pool) FindVolume(logger lager.Logger, teamID int, handle string) (Volume, bool, error) {
	worker, found, err := pool.provider.FindWorkerForVolume(
		logger.Session("find-worker"),
		teamID,
		handle,
	)
	if err != nil {
		return nil, false, err
	}

	if !found {
		return nil, false, nil
	}

	return worker.LookupVolume(logger, handle)
}

func (pool *pool) CreateVolume(logger lager.Logger, volumeSpec VolumeSpec, workerSpec WorkerSpec, volumeType db.VolumeType) (Volume, error) {
	worker, err := pool.chooseRandomWorkerForVolume(logger, workerSpec)
	if err != nil {
		return nil, err
	}

	return worker.CreateVolume(logger, volumeSpec, workerSpec.TeamID, volumeType)
}

func (pool *pool) ContainerInWorker(logger lager.Logger, owner db.ContainerOwner, workerSpec WorkerSpec) (bool, error) {
	workersWithContainer, err := pool.provider.FindWorkersForContainerByOwner(
		logger.Session("find-worker"),
		owner,
	)
	if err != nil {
		return false, err
	}

	compatibleWorkers, err := pool.allSatisfying(logger, workerSpec)
	if err != nil {
		return false, err
	}

	for _, w := range workersWithContainer {
		for _, c := range compatibleWorkers {
			if w.Name() == c.Name() {
				return true, nil
			}
		}
	}

	return false, nil
}

func (pool *pool) SelectWorker(
	ctx context.Context,
	owner db.ContainerOwner,
	containerSpec ContainerSpec,
	workerSpec WorkerSpec,
	strategy ContainerPlacementStrategy,
) (Client, error) {
	logger := lagerctx.FromContext(ctx)

	workersWithContainer, err := pool.provider.FindWorkersForContainerByOwner(
		logger.Session("find-worker"),
		owner,
	)
	if err != nil {
		return nil, err
	}

	compatibleWorkers, err := pool.allSatisfying(logger, workerSpec)
	if err != nil {
		return nil, err
	}

	for _, w := range workersWithContainer {
		for _, c := range compatibleWorkers {
			if w.Name() == c.Name() {
				return NewClient(c), nil
			}
		}
	}

	if !pool.policyChecker.ShouldCheckAction(policy.ActionSelectWorker) {
		selectedWorker, err := strategy.Choose(logger, compatibleWorkers, containerSpec)
		if err != nil {
			return nil, err
		}
		return NewClient(selectedWorker), nil
	}

	allowedWorkers := map[string]Worker{}
	for _, worker := range compatibleWorkers {
		allowedWorkers[worker.Name()] = worker
	}

	// Choose a worker and check with policy server if that worker is allowed
	// If not, remove it from the list of allowed workers and choose another
	// Go until a valid one is found or the list of allowed workers is empty
	reasons := []string{}
	for len(allowedWorkers) > 0 {
		allowedWorkerSlice := []Worker{}
		for _, val := range allowedWorkers {
			allowedWorkerSlice = append(allowedWorkerSlice, val)
		}
		selectedWorker, err := strategy.Choose(logger, allowedWorkerSlice, containerSpec)
		if err != nil {
			return nil, err
		}

		// TODO: include Job
		team, _ := ctx.Value("team").(string)
		pipeline, _ := ctx.Value("pipeline").(string)
		checkInput := policy.PolicyCheckInput{
			Action:   policy.ActionSelectWorker,
			Team:     team,
			Pipeline: pipeline,
			Data:     workerPolicyCheckData(selectedWorker, workerSpec, containerSpec),
		}
		result, err := pool.policyChecker.Check(checkInput)
		if err != nil {
			return nil, err
		}

		if err == nil && result.Allowed {
			return NewClient(selectedWorker), nil
		}

		reasons = append(reasons, result.Reasons...)
		delete(allowedWorkers, selectedWorker.Name())
	}

	return nil, NoValidWorkersError{
		Reasons: reasons,
	}
}

func (pool *pool) chooseRandomWorkerForVolume(
	logger lager.Logger,
	workerSpec WorkerSpec,
) (Worker, error) {
	workers, err := pool.allSatisfying(logger, workerSpec)
	if err != nil {
		return nil, err
	}

	return workers[rand.Intn(len(workers))], nil
}

func workerPolicyCheckData(worker Worker, workerSpec WorkerSpec, containerSpec ContainerSpec) map[string]interface{} {
	activeTasks, _ := worker.ActiveTasks()
	return map[string]interface{}{
		"selected_worker": map[string]interface{}{
			"name":             worker.Name(),
			"description":      worker.Description(),
			"build_containers": worker.BuildContainers(),
			"tags":             worker.Tags(),
			"uptime":           worker.Uptime(),
			"is_owned_by_team": worker.IsOwnedByTeam(),
			"ephemeral":        worker.Ephemeral(),
			"active_tasks":     activeTasks,
		},
		"worker_spec":    workerSpec,
		"container_spec": containerSpec,
	}
}
