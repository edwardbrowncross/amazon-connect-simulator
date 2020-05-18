package module

import (
	"fmt"

	"github.com/edwardbrowncross/amazon-connect-simulator/event"
	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

type setAttributes flow.Module

type setAttributesParams struct {
	Attribute []flow.KeyValue
}

func (m setAttributes) Run(call CallConnector) (next *flow.ModuleID, err error) {
	if m.Type != flow.ModuleSetAttributes {
		return nil, fmt.Errorf("module of type %s being run as setAttributes", m.Type)
	}
	p := setAttributesParams{}
	err = parameterResolver{call}.unmarshal(m.Parameters, &p)
	if err != nil {
		return
	}
	call.Emit(event.NewModuleEvent(flow.Module(m)))
	for _, a := range p.Attribute {
		call.SetContactData(a.K, a.V)
	}
	return m.Branches.GetLink(flow.BranchSuccess), nil
}
