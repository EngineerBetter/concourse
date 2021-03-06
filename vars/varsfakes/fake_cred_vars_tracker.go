// Code generated by counterfeiter. DO NOT EDIT.
package varsfakes

import (
	"sync"

	"github.com/concourse/concourse/vars"
)

type FakeCredVarsTracker struct {
	AddLocalVarStub        func(string, interface{}, bool)
	addLocalVarMutex       sync.RWMutex
	addLocalVarArgsForCall []struct {
		arg1 string
		arg2 interface{}
		arg3 bool
	}
	EnabledStub        func() bool
	enabledMutex       sync.RWMutex
	enabledArgsForCall []struct {
	}
	enabledReturns struct {
		result1 bool
	}
	enabledReturnsOnCall map[int]struct {
		result1 bool
	}
	GetStub        func(vars.VariableDefinition) (interface{}, bool, error)
	getMutex       sync.RWMutex
	getArgsForCall []struct {
		arg1 vars.VariableDefinition
	}
	getReturns struct {
		result1 interface{}
		result2 bool
		result3 error
	}
	getReturnsOnCall map[int]struct {
		result1 interface{}
		result2 bool
		result3 error
	}
	IterateInterpolatedCredsStub        func(vars.CredVarsTrackerIterator)
	iterateInterpolatedCredsMutex       sync.RWMutex
	iterateInterpolatedCredsArgsForCall []struct {
		arg1 vars.CredVarsTrackerIterator
	}
	ListStub        func() ([]vars.VariableDefinition, error)
	listMutex       sync.RWMutex
	listArgsForCall []struct {
	}
	listReturns struct {
		result1 []vars.VariableDefinition
		result2 error
	}
	listReturnsOnCall map[int]struct {
		result1 []vars.VariableDefinition
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeCredVarsTracker) AddLocalVar(arg1 string, arg2 interface{}, arg3 bool) {
	fake.addLocalVarMutex.Lock()
	fake.addLocalVarArgsForCall = append(fake.addLocalVarArgsForCall, struct {
		arg1 string
		arg2 interface{}
		arg3 bool
	}{arg1, arg2, arg3})
	fake.recordInvocation("AddLocalVar", []interface{}{arg1, arg2, arg3})
	fake.addLocalVarMutex.Unlock()
	if fake.AddLocalVarStub != nil {
		fake.AddLocalVarStub(arg1, arg2, arg3)
	}
}

func (fake *FakeCredVarsTracker) AddLocalVarCallCount() int {
	fake.addLocalVarMutex.RLock()
	defer fake.addLocalVarMutex.RUnlock()
	return len(fake.addLocalVarArgsForCall)
}

func (fake *FakeCredVarsTracker) AddLocalVarCalls(stub func(string, interface{}, bool)) {
	fake.addLocalVarMutex.Lock()
	defer fake.addLocalVarMutex.Unlock()
	fake.AddLocalVarStub = stub
}

func (fake *FakeCredVarsTracker) AddLocalVarArgsForCall(i int) (string, interface{}, bool) {
	fake.addLocalVarMutex.RLock()
	defer fake.addLocalVarMutex.RUnlock()
	argsForCall := fake.addLocalVarArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeCredVarsTracker) Enabled() bool {
	fake.enabledMutex.Lock()
	ret, specificReturn := fake.enabledReturnsOnCall[len(fake.enabledArgsForCall)]
	fake.enabledArgsForCall = append(fake.enabledArgsForCall, struct {
	}{})
	fake.recordInvocation("Enabled", []interface{}{})
	fake.enabledMutex.Unlock()
	if fake.EnabledStub != nil {
		return fake.EnabledStub()
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.enabledReturns
	return fakeReturns.result1
}

func (fake *FakeCredVarsTracker) EnabledCallCount() int {
	fake.enabledMutex.RLock()
	defer fake.enabledMutex.RUnlock()
	return len(fake.enabledArgsForCall)
}

func (fake *FakeCredVarsTracker) EnabledCalls(stub func() bool) {
	fake.enabledMutex.Lock()
	defer fake.enabledMutex.Unlock()
	fake.EnabledStub = stub
}

func (fake *FakeCredVarsTracker) EnabledReturns(result1 bool) {
	fake.enabledMutex.Lock()
	defer fake.enabledMutex.Unlock()
	fake.EnabledStub = nil
	fake.enabledReturns = struct {
		result1 bool
	}{result1}
}

func (fake *FakeCredVarsTracker) EnabledReturnsOnCall(i int, result1 bool) {
	fake.enabledMutex.Lock()
	defer fake.enabledMutex.Unlock()
	fake.EnabledStub = nil
	if fake.enabledReturnsOnCall == nil {
		fake.enabledReturnsOnCall = make(map[int]struct {
			result1 bool
		})
	}
	fake.enabledReturnsOnCall[i] = struct {
		result1 bool
	}{result1}
}

func (fake *FakeCredVarsTracker) Get(arg1 vars.VariableDefinition) (interface{}, bool, error) {
	fake.getMutex.Lock()
	ret, specificReturn := fake.getReturnsOnCall[len(fake.getArgsForCall)]
	fake.getArgsForCall = append(fake.getArgsForCall, struct {
		arg1 vars.VariableDefinition
	}{arg1})
	fake.recordInvocation("Get", []interface{}{arg1})
	fake.getMutex.Unlock()
	if fake.GetStub != nil {
		return fake.GetStub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2, ret.result3
	}
	fakeReturns := fake.getReturns
	return fakeReturns.result1, fakeReturns.result2, fakeReturns.result3
}

func (fake *FakeCredVarsTracker) GetCallCount() int {
	fake.getMutex.RLock()
	defer fake.getMutex.RUnlock()
	return len(fake.getArgsForCall)
}

func (fake *FakeCredVarsTracker) GetCalls(stub func(vars.VariableDefinition) (interface{}, bool, error)) {
	fake.getMutex.Lock()
	defer fake.getMutex.Unlock()
	fake.GetStub = stub
}

func (fake *FakeCredVarsTracker) GetArgsForCall(i int) vars.VariableDefinition {
	fake.getMutex.RLock()
	defer fake.getMutex.RUnlock()
	argsForCall := fake.getArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeCredVarsTracker) GetReturns(result1 interface{}, result2 bool, result3 error) {
	fake.getMutex.Lock()
	defer fake.getMutex.Unlock()
	fake.GetStub = nil
	fake.getReturns = struct {
		result1 interface{}
		result2 bool
		result3 error
	}{result1, result2, result3}
}

func (fake *FakeCredVarsTracker) GetReturnsOnCall(i int, result1 interface{}, result2 bool, result3 error) {
	fake.getMutex.Lock()
	defer fake.getMutex.Unlock()
	fake.GetStub = nil
	if fake.getReturnsOnCall == nil {
		fake.getReturnsOnCall = make(map[int]struct {
			result1 interface{}
			result2 bool
			result3 error
		})
	}
	fake.getReturnsOnCall[i] = struct {
		result1 interface{}
		result2 bool
		result3 error
	}{result1, result2, result3}
}

func (fake *FakeCredVarsTracker) IterateInterpolatedCreds(arg1 vars.CredVarsTrackerIterator) {
	fake.iterateInterpolatedCredsMutex.Lock()
	fake.iterateInterpolatedCredsArgsForCall = append(fake.iterateInterpolatedCredsArgsForCall, struct {
		arg1 vars.CredVarsTrackerIterator
	}{arg1})
	fake.recordInvocation("IterateInterpolatedCreds", []interface{}{arg1})
	fake.iterateInterpolatedCredsMutex.Unlock()
	if fake.IterateInterpolatedCredsStub != nil {
		fake.IterateInterpolatedCredsStub(arg1)
	}
}

func (fake *FakeCredVarsTracker) IterateInterpolatedCredsCallCount() int {
	fake.iterateInterpolatedCredsMutex.RLock()
	defer fake.iterateInterpolatedCredsMutex.RUnlock()
	return len(fake.iterateInterpolatedCredsArgsForCall)
}

func (fake *FakeCredVarsTracker) IterateInterpolatedCredsCalls(stub func(vars.CredVarsTrackerIterator)) {
	fake.iterateInterpolatedCredsMutex.Lock()
	defer fake.iterateInterpolatedCredsMutex.Unlock()
	fake.IterateInterpolatedCredsStub = stub
}

func (fake *FakeCredVarsTracker) IterateInterpolatedCredsArgsForCall(i int) vars.CredVarsTrackerIterator {
	fake.iterateInterpolatedCredsMutex.RLock()
	defer fake.iterateInterpolatedCredsMutex.RUnlock()
	argsForCall := fake.iterateInterpolatedCredsArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeCredVarsTracker) List() ([]vars.VariableDefinition, error) {
	fake.listMutex.Lock()
	ret, specificReturn := fake.listReturnsOnCall[len(fake.listArgsForCall)]
	fake.listArgsForCall = append(fake.listArgsForCall, struct {
	}{})
	fake.recordInvocation("List", []interface{}{})
	fake.listMutex.Unlock()
	if fake.ListStub != nil {
		return fake.ListStub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.listReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeCredVarsTracker) ListCallCount() int {
	fake.listMutex.RLock()
	defer fake.listMutex.RUnlock()
	return len(fake.listArgsForCall)
}

func (fake *FakeCredVarsTracker) ListCalls(stub func() ([]vars.VariableDefinition, error)) {
	fake.listMutex.Lock()
	defer fake.listMutex.Unlock()
	fake.ListStub = stub
}

func (fake *FakeCredVarsTracker) ListReturns(result1 []vars.VariableDefinition, result2 error) {
	fake.listMutex.Lock()
	defer fake.listMutex.Unlock()
	fake.ListStub = nil
	fake.listReturns = struct {
		result1 []vars.VariableDefinition
		result2 error
	}{result1, result2}
}

func (fake *FakeCredVarsTracker) ListReturnsOnCall(i int, result1 []vars.VariableDefinition, result2 error) {
	fake.listMutex.Lock()
	defer fake.listMutex.Unlock()
	fake.ListStub = nil
	if fake.listReturnsOnCall == nil {
		fake.listReturnsOnCall = make(map[int]struct {
			result1 []vars.VariableDefinition
			result2 error
		})
	}
	fake.listReturnsOnCall[i] = struct {
		result1 []vars.VariableDefinition
		result2 error
	}{result1, result2}
}

func (fake *FakeCredVarsTracker) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.addLocalVarMutex.RLock()
	defer fake.addLocalVarMutex.RUnlock()
	fake.enabledMutex.RLock()
	defer fake.enabledMutex.RUnlock()
	fake.getMutex.RLock()
	defer fake.getMutex.RUnlock()
	fake.iterateInterpolatedCredsMutex.RLock()
	defer fake.iterateInterpolatedCredsMutex.RUnlock()
	fake.listMutex.RLock()
	defer fake.listMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeCredVarsTracker) recordInvocation(key string, args []interface{}) {
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

var _ vars.CredVarsTracker = new(FakeCredVarsTracker)
