package module

import (
	"fmt"

	"github.com/edwardbrowncross/amazon-connect-simulator/event"
	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

type disconnect flow.Module

func (m disconnect) Run(call CallConnector) (next *flow.ModuleID, err error) {
	if m.Type != flow.ModuleDisconnect {
		return nil, fmt.Errorf("module of type %s being run as disconnect", m.Type)
	}
	call.Emit(event.NewModuleEvent(flow.Module(m)))
	call.Emit(event.DisconnectEvent{})
	return nil, nil
}
