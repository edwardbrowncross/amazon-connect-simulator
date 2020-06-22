package module

import (
	"errors"
	"fmt"

	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

type checkHoursOfOperation flow.Module

func (m checkHoursOfOperation) Run(call CallConnector) (next *flow.ModuleID, err error) {
	if m.Type != flow.ModuleCheckHoursOfOperation {
		return nil, fmt.Errorf("module of type %s being run as checkHoursOfOperation", m.Type)
	}

	h, ok := m.Parameters.Get("Hours")
	q := call.GetSystem(flow.SystemQueueName)

	var inHours bool
	var ihErr error
	if ok {
		inHours, ihErr = call.IsInHours(h.ResourceName, false)
	} else if q != nil {
		inHours, ihErr = call.IsInHours(*q, true)
	} else {
		ihErr = errors.New("Queue not set")
	}

	if ihErr != nil {
		return m.Branches.GetLink(flow.BranchError), nil
	}

	if !inHours {
		return m.Branches.GetLink(flow.BranchFalse), nil
	}
	return m.Branches.GetLink(flow.BranchTrue), nil
}
