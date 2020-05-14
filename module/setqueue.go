package module

import (
	"errors"
	"fmt"

	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

type setQueue flow.Module

func (m setQueue) Run(ctx CallContext) (next *flow.ModuleID, err error) {
	if m.Type != flow.ModuleSetQueue {
		return nil, fmt.Errorf("module of type %s being run as setQueue", m.Type)
	}
	p := m.Parameters.Get("Queue")
	if p == nil {
		return nil, errors.New("missing Queue parameter")
	}
	ctx.SetSystem(flow.SystemQueueARN, p.Value.(string))
	ctx.SetSystem(flow.SystemQueueName, p.ResourceName)
	return m.Branches.GetLink(flow.BranchSuccess), nil
}
