package module

import (
	"fmt"
	"strconv"
	"time"

	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

type getUserInput flow.Module

type getUserInputParams struct {
	Text             string
	Timeout          string
	MaxDigits        string
	TextToSpeechType string
}

func (m getUserInput) Run(call CallConnector) (next *flow.ModuleID, err error) {
	if m.Type != flow.ModuleGetUserInput {
		return nil, fmt.Errorf("module of type %s being run as getUserInput", m.Type)
	}
	pr := parameterResolver{call}
	p := getUserInputParams{}
	err = pr.unmarshal(m.Parameters, &p)
	if err != nil {
		return
	}
	if p.Text == "" {
		return m.Branches.GetLink(flow.BranchError), nil
	}
	txt := pr.jsonPath(p.Text)
	call.Send(txt, p.TextToSpeechType == "ssml")
	md, err := strconv.Atoi(p.MaxDigits)
	if err != nil {
		return nil, fmt.Errorf("invalid MaxDigits: %s", p.MaxDigits)
	}
	tm, err := strconv.Atoi(p.Timeout)
	if err != nil {
		return nil, fmt.Errorf("invalid Timeout: %s", p.Timeout)
	}
	in, ok := call.Receive(md, time.Duration(tm)*time.Second, rune(0))
	if !ok {
		return m.Branches.GetLink(flow.BranchTimeout), nil
	}
	return evaluateConditions(m.Branches, in)
}
