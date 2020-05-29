package module

import (
	"fmt"

	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

type setVoice flow.Module

type setVoiceParams struct {
	GlobalVoice string
}

func (m setVoice) Run(call CallConnector) (next *flow.ModuleID, err error) {
	if m.Type != flow.ModuleSetVoice {
		return nil, fmt.Errorf("module of type %s being run as setVoice", m.Type)
	}
	pr := parameterResolver{call}
	var p setVoiceParams
	err = pr.unmarshal(m.Parameters, &p)
	if err != nil {
		return
	}
	call.SetSystem(flow.SystemTextToSpeechVoice, p.GlobalVoice)
	return m.Branches.GetLink(flow.BranchSuccess), nil
}
