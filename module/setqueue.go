package module

import (
	"errors"
	"fmt"

	"github.com/edwardbrowncross/amazon-connect-simulator/event"
	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

type setQueue flow.Module

func (m setQueue) Run(call CallConnector) (next *flow.ModuleID, err error) {
	if m.Type != flow.ModuleSetQueue {
		return nil, fmt.Errorf("module of type %s being run as setQueue", m.Type)
	}
	p := m.Parameters.Get("Queue")
	if p == nil {
		return nil, errors.New("missing Queue parameter")
	}
	call.Emit(event.NewModuleEvent(flow.Module(m)))
	call.SetSystem(flow.SystemQueueARN, p.Value.(string))
	call.SetSystem(flow.SystemQueueName, p.ResourceName)
	return m.Branches.GetLink(flow.BranchSuccess), nil
}
