package module

import (
	"errors"
	"fmt"

	"github.com/edwardbrowncross/amazon-connect-simulator/event"
	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

type transfer flow.Module

func (m transfer) Run(call CallConnector) (next *flow.ModuleID, err error) {
	if m.Type != flow.ModuleTransfer {
		return nil, fmt.Errorf("module of type %s being run as transfer", m.Type)
	}
	switch m.Target {
	case flow.TargetFlow:
		cfid := m.Parameters.Get("ContactFlowId")
		if cfid == nil {
			return nil, errors.New("missing ContactFlowId parameter")
		}
		call.Emit(event.NewModuleEvent(flow.Module(m)))
		next = call.GetFlowStart(cfid.ResourceName)
		if next == nil {
			return m.Branches.GetLink(flow.BranchError), nil
		}
		call.Emit(event.FlowTransferEvent{FlowARN: cfid.Value.(string), FlowName: cfid.ResourceName})
		return
	case flow.TargetQueue:
		call.Emit(event.NewModuleEvent(flow.Module(m)))
		queue := call.GetSystem(string(flow.SystemQueueName))
		if qs, ok := queue.(string); !ok || qs == "" {
			return m.Branches.GetLink(flow.BranchError), nil
		}
		arn := call.GetSystem(string(flow.SystemQueueARN))
		call.Emit(event.QueueTransferEvent{QueueARN: arn.(string), QueueName: queue.(string)})
		return nil, nil
	default:
		return nil, fmt.Errorf("unhandled transfer target: %s", m.Target)
	}
}
