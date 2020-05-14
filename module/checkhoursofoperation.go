package module

import (
	"fmt"

	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

type checkHoursOfOperation flow.Module

func (m checkHoursOfOperation) Run(ctx CallContext) (next *flow.ModuleID, err error) {
	if m.Type != flow.ModuleCheckHoursOfOperation {
		return nil, fmt.Errorf("module of type %s being run as checkHoursOfOperation", m.Type)
	}
	return m.Branches.GetLink(flow.BranchTrue), nil
}
