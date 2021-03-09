// Code generated by counterfeiter. DO NOT EDIT.
package workerfakes

import (
	"sync"

	"code.cloudfoundry.org/lager"
	"github.com/concourse/concourse/atc/worker"
)

type FakePoolCallbacks struct {
	SelectedWorkerStub        func(lager.Logger, worker.Worker)
	selectedWorkerMutex       sync.RWMutex
	selectedWorkerArgsForCall []struct {
		arg1 lager.Logger
		arg2 worker.Worker
	}
	WaitingForWorkerStub        func(lager.Logger)
	waitingForWorkerMutex       sync.RWMutex
	waitingForWorkerArgsForCall []struct {
		arg1 lager.Logger
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakePoolCallbacks) SelectedWorker(arg1 lager.Logger, arg2 worker.Worker) {
	fake.selectedWorkerMutex.Lock()
	fake.selectedWorkerArgsForCall = append(fake.selectedWorkerArgsForCall, struct {
		arg1 lager.Logger
		arg2 worker.Worker
	}{arg1, arg2})
	stub := fake.SelectedWorkerStub
	fake.recordInvocation("SelectedWorker", []interface{}{arg1, arg2})
	fake.selectedWorkerMutex.Unlock()
	if stub != nil {
		fake.SelectedWorkerStub(arg1, arg2)
	}
}

func (fake *FakePoolCallbacks) SelectedWorkerCallCount() int {
	fake.selectedWorkerMutex.RLock()
	defer fake.selectedWorkerMutex.RUnlock()
	return len(fake.selectedWorkerArgsForCall)
}

func (fake *FakePoolCallbacks) SelectedWorkerCalls(stub func(lager.Logger, worker.Worker)) {
	fake.selectedWorkerMutex.Lock()
	defer fake.selectedWorkerMutex.Unlock()
	fake.SelectedWorkerStub = stub
}

func (fake *FakePoolCallbacks) SelectedWorkerArgsForCall(i int) (lager.Logger, worker.Worker) {
	fake.selectedWorkerMutex.RLock()
	defer fake.selectedWorkerMutex.RUnlock()
	argsForCall := fake.selectedWorkerArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakePoolCallbacks) WaitingForWorker(arg1 lager.Logger) {
	fake.waitingForWorkerMutex.Lock()
	fake.waitingForWorkerArgsForCall = append(fake.waitingForWorkerArgsForCall, struct {
		arg1 lager.Logger
	}{arg1})
	stub := fake.WaitingForWorkerStub
	fake.recordInvocation("WaitingForWorker", []interface{}{arg1})
	fake.waitingForWorkerMutex.Unlock()
	if stub != nil {
		fake.WaitingForWorkerStub(arg1)
	}
}

func (fake *FakePoolCallbacks) WaitingForWorkerCallCount() int {
	fake.waitingForWorkerMutex.RLock()
	defer fake.waitingForWorkerMutex.RUnlock()
	return len(fake.waitingForWorkerArgsForCall)
}

func (fake *FakePoolCallbacks) WaitingForWorkerCalls(stub func(lager.Logger)) {
	fake.waitingForWorkerMutex.Lock()
	defer fake.waitingForWorkerMutex.Unlock()
	fake.WaitingForWorkerStub = stub
}

func (fake *FakePoolCallbacks) WaitingForWorkerArgsForCall(i int) lager.Logger {
	fake.waitingForWorkerMutex.RLock()
	defer fake.waitingForWorkerMutex.RUnlock()
	argsForCall := fake.waitingForWorkerArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakePoolCallbacks) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.selectedWorkerMutex.RLock()
	defer fake.selectedWorkerMutex.RUnlock()
	fake.waitingForWorkerMutex.RLock()
	defer fake.waitingForWorkerMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakePoolCallbacks) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ worker.PoolCallbacks = new(FakePoolCallbacks)
