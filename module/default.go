package module

import (
	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

type passthrough flow.Module

func (m passthrough) Run(call CallConnector) (next *flow.ModuleID, err error) {
	return m.Branches.GetLink(flow.BranchSuccess), nil
}
