package module

import (
	"github.com/edwardbrowncross/amazon-connect-simulator/call"
	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

func MakeRunner(m flow.Module) call.Runner {
	switch m.Type {
	case flow.ModuleStoreUserInput:
		return storeUserInput(m)
	case flow.ModuleCheckAttribute:
		return checkAttribute(m)
	case flow.ModuleTransfer:
		return transfer(m)
	case flow.ModulePlayPrompt:
		return playPrompt(m)
	case flow.ModuleDisconnect:
		return disconnect(m)
	case flow.ModuleSetQueue:
		return setQueue(m)
	case flow.ModuleGetUserInput:
		return getUserInput(m)
	case flow.ModuleSetAttributes:
		return setAttributes(m)
	case flow.ModuleInvokeExternalResource:
		return invokeExternalResource(m)
	case flow.ModuleCheckHoursOfOperation:
		return checkHoursOfOperation(m)
	default:
		return passthrough(m)
	}
}
