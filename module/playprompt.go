package module

import (
	"fmt"

	"github.com/edwardbrowncross/amazon-connect-simulator/event"
	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

type playPrompt flow.Module

type playPromptParams struct {
	Text             string
	TextToSpeechType string
}

func (m playPrompt) Run(call CallConnector) (next *flow.ModuleID, err error) {
	if m.Type != flow.ModulePlayPrompt {
		return nil, fmt.Errorf("module of type %s being run as playPrompt", m.Type)
	}
	pr := parameterResolver{call}
	p := playPromptParams{}
	err = pr.unmarshal(m.Parameters, &p)
	if err != nil {
		return
	}
	call.Emit(event.NewModuleEvent(flow.Module(m)))
	txt, err := pr.jsonPath(p.Text)
	if err != nil {
		return
	}
	call.Emit(event.PromptEvent{
		Text: txt,
		SSML: p.TextToSpeechType == "ssml",
	})
	call.Send(txt)
	return m.Branches.GetLink(flow.BranchSuccess), nil
}
