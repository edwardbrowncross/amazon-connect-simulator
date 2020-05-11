package simulator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/edwardbrowncross/amazon-connect-simulator/call"
	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
	"github.com/edwardbrowncross/amazon-connect-simulator/module"
)

type CallSimulator struct {
	lambdas      map[string]interface{}
	flows        map[string]flow.Flow
	modules      map[flow.ModuleID]flow.Module
	startingFlow *flow.Flow
}

func NewCallSimulator() CallSimulator {
	return CallSimulator{
		lambdas: map[string]interface{}{},
		flows:   map[string]flow.Flow{},
		modules: map[flow.ModuleID]flow.Module{},
	}
}

func (cs *CallSimulator) LoadFlow(flow flow.Flow) {
	cs.flows[flow.Metadata.Name] = flow
	for _, m := range flow.Modules {
		cs.modules[m.ID] = m
	}
}

func (cs *CallSimulator) LoadFlowJSON(bytes []byte) error {
	f := flow.Flow{}
	err := json.Unmarshal(bytes, &f)
	if err != nil {
		return err
	}
	cs.LoadFlow(f)
	return nil
}

func (cs *CallSimulator) RegisterLambda(name string, fn interface{}) error {
	err := validateLambda(fn)
	if err != nil {
		return err
	}
	cs.lambdas[name] = fn
	return nil
}

func (cs *CallSimulator) SetStartingFlow(flowName string) error {
	f, ok := cs.flows[flowName]
	if !ok {
		return errors.New("starting flow not found. Load the flow with LoadFlow before calling this method")
	}
	cs.startingFlow = &f
	return nil
}

func (cs *CallSimulator) StartCall(config call.Config) (call.Call, error) {
	if cs.startingFlow == nil {
		return call.Call{}, errors.New("no starting flow set. Call SetStartingFlow before starting a call")
	}
	return call.New(config, call.FlowDescriber{
		GetLambda:    cs.lookupLambda,
		GetFlowStart: cs.getFlowStart,
		GetRunner:    cs.getRunner,
	}, *&cs.startingFlow.Start), nil
}

func (cs *CallSimulator) lookupLambda(arn string) interface{} {
	for k, v := range cs.lambdas {
		if strings.Contains(arn, k) {
			return v
		}
	}
	return nil
}

func (cs *CallSimulator) getFlowStart(flowName string) *flow.ModuleID {
	f, ok := cs.flows[flowName]
	if !ok {
		return nil
	}
	return &f.Start
}

func (cs *CallSimulator) getRunner(moduleID flow.ModuleID) call.Runner {
	m, ok := cs.modules[moduleID]
	if !ok {
		return nil
	}
	return module.MakeRunner(m)
}

func validateLambda(fn interface{}) error {
	fnt := reflect.TypeOf(fn)
	if fnt.Kind() != reflect.Func {
		return fmt.Errorf("wanted function but got %s", fnt.Kind())
	}
	if fnt.NumIn() != 2 {
		return fmt.Errorf("expected function to take 2 parameters but it takes %d", fnt.NumIn())
	}
	contextt := reflect.TypeOf((*context.Context)(nil)).Elem()
	if !fnt.In(0).Implements(contextt) {
		return errors.New("expected first argument to be a context.Context")
	}
	if fnt.In(1).Kind() != reflect.Struct {
		return fmt.Errorf("expected second argument to be struct but it was %s", fnt.In(1).Kind())
	}
	if fnt.NumOut() != 2 {
		return fmt.Errorf("expected function to return 2 elements but it returns %d", fnt.NumOut())
	}
	if fnt.Out(0).Kind() != reflect.Struct {
		return fmt.Errorf("expected first return to be struct but it was %s", fnt.Out(0).Kind())
	}
	errort := reflect.TypeOf((*error)(nil)).Elem()
	if !fnt.Out(1).Implements(errort) {
		return errors.New("expected seclnd return to be an error")
	}
	return nil
}
