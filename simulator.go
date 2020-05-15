package simulator

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/edwardbrowncross/amazon-connect-simulator/call"
	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
	"github.com/edwardbrowncross/amazon-connect-simulator/module"
)

// CallSimulator is capable of starting new simulated call flows.
type CallSimulator struct {
	lambdas      map[string]interface{}
	flows        map[string]flow.Flow
	modules      map[flow.ModuleID]flow.Module
	startingFlow *flow.Flow
}

// New creates a new call simulator.
// It is created blank and must be set up using its attached methods.
func New() CallSimulator {
	return CallSimulator{
		lambdas: map[string]interface{}{},
		flows:   map[string]flow.Flow{},
		modules: map[flow.ModuleID]flow.Module{},
	}
}

// LoadFlow loads an unmarshalled call flow into the simulator.
// Do this with all flows that form part of your call flows before starting a call.
func (cs *CallSimulator) LoadFlow(flow flow.Flow) {
	cs.flows[flow.Metadata.Name] = flow
	for _, m := range flow.Modules {
		cs.modules[m.ID] = m
	}
}

// LoadFlowJSON takes a byte array containing a json file exported from Amazon Connect.
// It does the same thing as LoadFlow, except that it does the unmarshalling for you.
func (cs *CallSimulator) LoadFlowJSON(bytes []byte) error {
	f := flow.Flow{}
	err := json.Unmarshal(bytes, &f)
	if err != nil {
		return err
	}
	cs.LoadFlow(f)
	return nil
}

// RegisterLambda specifies how external lambda invocations will be handled.
// name is a string that forms part of the lambda's ARN (such as its name).
// fn is function like handle(context.Context, struct) (struct, error). It will be passed an Amazon Connect lambda event.
// You must specify a function for each external lambda invocation before starting a simulated call.
func (cs *CallSimulator) RegisterLambda(name string, fn interface{}) error {
	err := module.ValidateLambda(fn)
	if err != nil {
		return err
	}
	cs.lambdas[name] = fn
	return nil
}

// SetStartingFlow specifies the name of the flow that should be run when a new call comes in.
// The name is the full name given to the flow in the Amazon Connect ui.
// You must run this once before starting a simulated call.
func (cs *CallSimulator) SetStartingFlow(flowName string) error {
	f, ok := cs.flows[flowName]
	if !ok {
		return errors.New("starting flow not found. Load the flow with LoadFlow before calling this method")
	}
	cs.startingFlow = &f
	return nil
}

// StartCall starts a new call asynchronously and returns a Call object for interacting with that call.
// Many independent calls can be spawned from one simulator.
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

func (cs *CallSimulator) getRunner(moduleID flow.ModuleID) module.Runner {
	m, ok := cs.modules[moduleID]
	if !ok {
		return nil
	}
	return module.MakeRunner(m)
}
