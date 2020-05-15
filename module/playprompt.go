package module

import (
	"fmt"

	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

type playPrompt flow.Module

type playPromptParams struct {
	Text string
}

func (m playPrompt) Run(call CallConnector) (next *flow.ModuleID, err error) {
	if m.Type != flow.ModulePlayPrompt {
		return nil, fmt.Errorf("module of type %s being run as playPrompt", m.Type)
	}
	p := playPromptParams{}
	err = parameterResolver{call}.unmarshal(m.Parameters, &p)
	if err != nil {
		return
	}
	call.Send(p.Text)
	return m.Branches.GetLink(flow.BranchSuccess), nil
}
