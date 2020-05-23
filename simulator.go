package simulator

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
	"github.com/edwardbrowncross/amazon-connect-simulator/module"
)

// Simulator is capable of starting new simulated call flows.
type Simulator struct {
	lambdas      map[string]interface{}
	flows        map[string]flow.Flow
	modules      map[flow.ModuleID]flow.Module
	startingFlow *flow.Flow
}

// New creates a new call simulator.
// It is created blank and must be set up using its attached methods.
func New() Simulator {
	return Simulator{
		lambdas: map[string]interface{}{},
		flows:   map[string]flow.Flow{},
		modules: map[flow.ModuleID]flow.Module{},
	}
}

// LoadFlow loads an unmarshalled call flow into the simulator.
// Do this with all flows that form part of your call flows before starting a call.
func (cs *Simulator) LoadFlow(f flow.Flow) {
	f = flow.Dedeprecate(f)
	cs.flows[f.Metadata.Name] = f
	for _, m := range f.Modules {
		cs.modules[m.ID] = m
	}
}

// LoadFlowJSON takes a byte array containing a json file exported from Amazon Connect.
// It does the same thing as LoadFlow, except that it does the unmarshalling for you.
func (cs *Simulator) LoadFlowJSON(bytes []byte) error {
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
func (cs *Simulator) RegisterLambda(name string, fn interface{}) error {
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
func (cs *Simulator) SetStartingFlow(flowName string) error {
	f, ok := cs.flows[flowName]
	if !ok {
		return errors.New("starting flow not found. Load the flow with LoadFlow before calling this method")
	}
	cs.startingFlow = &f
	return nil
}

// StartCall starts a new call asynchronously and returns a Call object for interacting with that call.
// Many independent calls can be spawned from one simulator.
func (cs *Simulator) StartCall(config CallConfig) (*Call, error) {
	if cs.startingFlow == nil {
		return nil, errors.New("no starting flow set. Call SetStartingFlow before starting a call")
	}
	return newCall(config, &simulatorConnector{cs}, *&cs.startingFlow.Start), nil
}

// simulatorConnector exposes methods for modules to get information from the base simulator.
type simulatorConnector struct {
	*Simulator
}

// GetLambda gets a lamda using a partial ARN match.
func (cs *simulatorConnector) GetLambda(arn string) interface{} {
	for k, v := range cs.lambdas {
		if strings.Contains(arn, k) {
			return v
		}
	}
	return nil
}

// GetFlowStart gets the module ID of the block at the start of a flow with the given name.
func (cs *simulatorConnector) GetFlowStart(flowName string) *flow.ModuleID {
	f, ok := cs.flows[flowName]
	if !ok {
		return nil
	}
	return &f.Start
}

// GetRunner finds the block with the given ID and wraps in the module providing its functionality.
func (cs *simulatorConnector) GetRunner(moduleID flow.ModuleID) module.Runner {
	m, ok := cs.modules[moduleID]
	if !ok {
		return nil
	}
	return module.MakeRunner(m)
}
