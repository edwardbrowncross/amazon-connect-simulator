package module

import (
	"encoding/json"
	"testing"
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
	jsonOK := `{
		"id":"55c7b51c-ab55-4c63-ac42-235b4a0f904f",
		"type":"PlayPrompt",
		"branches":[{"condition":"Success","transition":"00000000-0000-4000-0000-000000000001"}],
		"parameters":[
			{"name":"Text","value":"Thanks for your call","namespace":null},
			{"name":"TextToSpeechType","value":"text"}
		]
	}`
	testCases := []struct {
		desc   string
		module string
		exp    string
		expErr string
		expOut string
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
			desc:   "success",
			module: jsonOK,
			exp:    "00000000-0000-4000-0000-000000000001",
			expOut: "Thanks for your call",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			var mod playPrompt
			err := json.Unmarshal([]byte(tC.module), &mod)
			if err != nil {
				t.Fatalf("unexpected error unmarshalling module: %v", err)
			}
			state := testCallState{}.init()
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
		})
	}
}
