package module

import (
	"fmt"

	"github.com/edwardbrowncross/amazon-connect-simulator/call"
	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

type playPrompt flow.Module

type playPromptParams struct {
	Text string
}

func (m playPrompt) Run(ctx *call.Context) (next *flow.ModuleID, err error) {
	if m.Type != flow.ModulePlayPrompt {
		return nil, fmt.Errorf("module of type %s being run as playPrompt", m.Type)
	}
	p := playPromptParams{}
	err = ctx.UnmarshalParameters(m.Parameters, &p)
	if err != nil {
		return
	}
	ctx.Send(p.Text)
	return m.Branches.GetLink(flow.BranchSuccess), nil
}
