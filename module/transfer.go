package module

import (
	"errors"
	"fmt"

	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

type transfer flow.Module

func (m transfer) Run(ctx CallContext) (next *flow.ModuleID, err error) {
	if m.Type != flow.ModuleTransfer {
		return nil, fmt.Errorf("module of type %s being run as transfer", m.Type)
	}
	switch m.Target {
	case flow.TargetFlow:
		cfid := m.Parameters.Get("ContactFlowId")
		if cfid == nil {
			return nil, errors.New("missing ContextFlowId parameter")
		}
		return ctx.GetFlowStart(cfid.ResourceName), nil
	case flow.TargetQueue:
		queue := ctx.GetSystem(string(flow.SystemQueueName))
		ctx.Send(fmt.Sprintf("<transfer to queue %s>", queue))
		return nil, nil
	default:
		return nil, fmt.Errorf("unhandled transfer target: %s", m.Target)
	}
}
