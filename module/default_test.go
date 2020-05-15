package module

import (
	"encoding/json"
	"testing"
)

func TestPassthrough(t *testing.T) {
	jsonOK := `{
		"id":"43dcc4f2-3392-4a38-90ed-0216f8594ea8",
		"type":"SetVoice",
		"branches":[
			{"condition":"Success", "transition":"00000000-0000-4000-0000-000000000001"}
		],
		"parameters":[{"name": "GlobalVoice", "value": "Joanna"}]
	}`
	testCases := []struct {
		desc   string
		module string
		exp    string
		expErr string
	}{
		{
			desc:   "success",
			module: jsonOK,
			exp:    "00000000-0000-4000-0000-000000000001",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			var mod passthrough
			err := json.Unmarshal([]byte(tC.module), &mod)
			if err != nil {
				t.Fatalf("unexpected error unmarshalling module: %v", err)
			}
			next, err := mod.Run(testContext{}.init())
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
		})
	}
}
