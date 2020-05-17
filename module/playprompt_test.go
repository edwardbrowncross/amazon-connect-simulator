package module

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/edwardbrowncross/amazon-connect-simulator/event"
)

func TestPlayPrompt(t *testing.T) {
	jsonBadMod := `{
		"id":"43dcc4f2-3392-4a38-90ed-0216f8594ea8",
		"type":"Transfer"
	}`
	jsonBadParam := `{
		"id":"43dcc4f2-3392-4a38-90ed-0216f8594ea8",
		"type":"PlayPrompt"
	}`
	jsonBadPath := `{
		"id":"55c7b51c-ab55-4c63-ac42-235b4a0f904f",
		"type":"PlayPrompt",
		"branches":[{"condition":"Success","transition":"00000000-0000-4000-0000-000000000001"}],
		"parameters":[{"name":"Text","value":"Thanks for your call $.Computer.name"}, {"name":"TextToSpeechType","value":"text"}]
	}`
	jsonOK := `{
		"id":"55c7b51c-ab55-4c63-ac42-235b4a0f904f",
		"type":"PlayPrompt",
		"branches":[{"condition":"Success","transition":"00000000-0000-4000-0000-000000000001"}],
		"parameters":[
			{"name":"Text","value":"Thanks for your call, $.Attributes.name.","namespace":null},
			{"name":"TextToSpeechType","value":"text"}
		]
	}`
	jsonOKSSML := `{
		"id":"55c7b51c-ab55-4c63-ac42-235b4a0f904f",
		"type":"PlayPrompt",
		"branches":[{"condition":"Success","transition":"00000000-0000-4000-0000-000000000001"}],
		"parameters":[
			{"name":"Text","value":"<speak>Thanks for your call.</speak>"},
			{"name":"TextToSpeechType","value":"ssml"}
		]
	}`
	testCases := []struct {
		desc   string
		module string
		exp    string
		expErr string
		expOut string
		expEvt []event.Event
	}{
		{
			desc:   "wrong module",
			module: jsonBadMod,
			expErr: "module of type Transfer being run as playPrompt",
		},
		{
			desc:   "missing parameter",
			module: jsonBadParam,
			expErr: "missing parameter Text",
		},
		{
			desc:   "bad JSON Path",
			module: jsonBadPath,
			expErr: "unknown namespace: Computer",
		},
		{
			desc:   "success",
			module: jsonOK,
			exp:    "00000000-0000-4000-0000-000000000001",
			expOut: "Thanks for your call, Edward.",
			expEvt: []event.Event{
				event.PromptEvent{Text: "Thanks for your call, Edward.", SSML: false},
			},
		},
		{
			desc:   "success - SSML",
			module: jsonOKSSML,
			exp:    "00000000-0000-4000-0000-000000000001",
			expOut: "<speak>Thanks for your call.</speak>",
			expEvt: []event.Event{
				event.PromptEvent{Text: "<speak>Thanks for your call.</speak>", SSML: true},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			var mod playPrompt
			err := json.Unmarshal([]byte(tC.module), &mod)
			if err != nil {
				t.Fatalf("unexpected error unmarshalling module: %v", err)
			}
			state := testCallState{
				contactData: map[string]string{
					"name": "Edward",
				},
			}.init()
			next, err := mod.Run(state)
			errStr := ""
			if err != nil {
				errStr = err.Error()
			}
			if errStr != tC.expErr {
				t.Errorf("expected error of '%s' but got '%s'", tC.expErr, errStr)
			}
			nextStr := ""
			if next != nil {
				nextStr = string(*next)
			}
			if nextStr != tC.exp {
				t.Errorf("expected next of '%s' but got '%v'", tC.exp, *next)
			}
			if state.o != tC.expOut {
				t.Errorf("expected ouptut of '%s' but got '%s'", tC.expOut, state.o)
			}
			if tC.expEvt != nil && !reflect.DeepEqual(tC.expEvt, state.events) {
				t.Errorf("expected events of '%v' but got '%v'", tC.expEvt, state.events)
			}
		})
	}
}
