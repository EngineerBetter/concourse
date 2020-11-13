package exec

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"reflect"
	"regexp"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagerctx"
	"github.com/concourse/concourse/atc/metric"
	"github.com/concourse/concourse/atc/worker"
	"github.com/concourse/concourse/atc/worker/transport"
)

type Retriable struct {
	Cause error
	Count int
}

type RetryableError string

func (r Retriable) Error() string {
	return fmt.Sprintf("retriable: attempt %d: %s", r.Count, r.Cause.Error())
}

type RetryErrorStep struct {
	Step

	delegateFactory BuildStepDelegateFactory
}

func RetryError(step Step, delegateFactory BuildStepDelegateFactory) Step {
	return RetryErrorStep{
		Step:            step,
		delegateFactory: delegateFactory,
	}
}

func (step RetryErrorStep) Run(ctx context.Context, state RunState) (bool, error) {
	logger := lagerctx.FromContext(ctx)
	runOk, runErr := step.Step.Run(ctx, state)

	// If the build has been aborted, then no need to retry.
	select {
	case <-ctx.Done():
		return runOk, runErr
	default:
	}

	if runErr != nil {
		if ok, errorType := step.toRetry(logger, runErr); ok {
			logger.Info("retriable", lager.Data{"error": runErr.Error()})
			delegate := step.delegateFactory.BuildStepDelegate(state)
			delegate.Errored(logger, fmt.Sprintf("%s, will retry ...", runErr.Error()))
			team, _ := ctx.Value("team").(string)
			retriedErrorsLabels := metric.RetriedErrorsLabels{RetryableError: errorType, TeamName: team}
			if _, ok := metric.Metrics.RetriedErrors[retriedErrorsLabels]; !ok {
				metric.Metrics.RetriedErrors[retriedErrorsLabels] = &metric.Counter{}
			}
			metric.Metrics.RetriedErrors[retriedErrorsLabels].Inc()
			var r Retriable
			if errors.As(runErr, &r) {
				runErr = Retriable{r.Cause, r.Count + 1}
			} else {
				runErr = Retriable{runErr, 0}
			}
		}
	}
	return runOk, runErr
}

func (step RetryErrorStep) toRetry(logger lager.Logger, err error) (bool, metric.RetryableError) {
	var urlError *url.Error
	var netError net.Error
	if errors.As(err, &transport.WorkerMissingError{}) || errors.As(err, &transport.WorkerUnreachableError{}) || errors.As(err, &urlError) {
		logger.Debug("retry-error",
			lager.Data{"err_type": reflect.TypeOf(err).String(), "err": err.Error()})
		return true, "worker_missing"
	} else if errors.As(err, &netError) || regexp.MustCompile(`worker .+ disappeared`).MatchString(err.Error()) {
		logger.Debug("retry-error",
			lager.Data{"err_type": reflect.TypeOf(err).String(), "err": err})
		return true, "worker_disappeared"
	} else if errors.As(err, &worker.ErrNoWorkers) || errors.Is(err, &worker.NoValidWorkersError{}) {
		logger.Debug("retry-error",
			lager.Data{"err_type": reflect.TypeOf(err).String(), "err": err})
		return true, "no_workers"
	}
	return false, ""
}
