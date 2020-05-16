package module

import (
	"fmt"
	"strconv"
	"time"

	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

type storeUserInput flow.Module

type storeUserInputParams struct {
	Text      string
	Timeout   string
	MaxDigits int
}

func (m storeUserInput) Run(call CallConnector) (next *flow.ModuleID, err error) {
	if m.Type != flow.ModuleStoreUserInput {
		return nil, fmt.Errorf("module of type %s being run as storeUserInput", m.Type)
	}
	pr := parameterResolver{call}
	p := storeUserInputParams{}
	err = pr.unmarshal(m.Parameters, &p)
	if err != nil {
		return
	}
	txt, err := pr.jsonPath(p.Text)
	if err != nil {
		return
	}
	call.Send(txt)
	timeout, err := strconv.Atoi(p.Timeout)
	if err != nil {
		return
	}
	entry := call.Receive(p.MaxDigits, time.Duration(timeout)*time.Second)
	if entry == nil {
		next = m.Branches.GetLink(flow.BranchError)
		return
	}
	call.SetSystem(string(flow.SystemLastUserInput), *entry)
	next = m.Branches.GetLink(flow.BranchSuccess)
	return
}
