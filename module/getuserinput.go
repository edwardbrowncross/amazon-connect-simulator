package module

import (
	"fmt"
	"strconv"
	"time"

	"github.com/edwardbrowncross/amazon-connect-simulator/call"
	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

type getUserInput flow.Module

type getUserInputParams struct {
	Text      string
	Timeout   string
	MaxDigits string
}

func (m getUserInput) Run(ctx *call.Context) (next *flow.ModuleID, err error) {
	if m.Type != flow.ModuleGetUserInput {
		return nil, fmt.Errorf("module of type %s being run as getUserInput", m.Type)
	}
	p := getUserInputParams{}
	err = ctx.UnmarshalParameters(m.Parameters, &p)
	if err != nil {
		return
	}
	ctx.Send(p.Text)
	md, err := strconv.Atoi(p.MaxDigits)
	if err != nil {
		return nil, fmt.Errorf("invalid MaxDigits: %s", p.MaxDigits)
	}
	tm, err := strconv.Atoi(p.Timeout)
	if err != nil {
		return nil, fmt.Errorf("invalid Timeout: %s", p.Timeout)
	}
	in := ctx.Receive(md, time.Duration(tm)*time.Second)
	if in == nil {
		return m.Branches.GetLink(flow.BranchTimeout), nil
	}
	return evaluateConditions(m.Branches, *in)
}
