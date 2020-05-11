package module

import (
	"fmt"

	"github.com/edwardbrowncross/amazon-connect-simulator/call"
	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

type disconnect flow.Module

func (m disconnect) Run(ctx *call.Context) (next *flow.ModuleID, err error) {
	if m.Type != flow.ModuleDisconnect {
		return nil, fmt.Errorf("module of type %s being run as disconnect", m.Type)
	}
	return nil, nil
}
