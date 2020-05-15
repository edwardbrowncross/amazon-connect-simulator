package module

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestSetAttributes(t *testing.T) {
	jsonBadMod := `{
		"id":"43dcc4f2-3392-4a38-90ed-0216f8594ea8",
		"type":"Transfer"
	}`
	jsonBadParam := `{
		"id":"43dcc4f2-3392-4a38-90ed-0216f8594ea8",
		"type":"SetAttributes",
		"parameters":[{"name":"Attribute","value":"securityPassed"}]
	}`
	jsonOK := `{
		"id":"55c7b51c-ab55-4c63-ac42-235b4a0f904f",
		"type":"SetAttributes",
		"branches":[
			{"condition":"Success","transition":"00000000-0000-4000-0000-000000000001"},
			{"condition":"Error","transition":"00000000-0000-4000-0000-000000000002"}
		],
		"parameters":[{"name":"Attribute","value":"securityPassed","key":"authorized","namespace":"External"}]
	}`
	testCases := []struct {
		desc    string
		module  string
		state   *testCallState
		exp     string
		expErr  string
		expAttr map[string]string
	}{
		{
			desc:   "wrong module",
			module: jsonBadMod,
			expErr: "module of type Transfer being run as setAttributes",
		},
		{
			desc:   "bad parameter",
			module: jsonBadParam,
			expErr: "type mismatch in field Attribute. Cannot convert string to flow.KeyValue",
		},
		{
			desc:   "success",
			module: jsonOK,
			state: testCallState{
				external: map[string]string{
					"securityPassed": "yes",
				},
			}.init(),
			exp: "00000000-0000-4000-0000-000000000001",
			expAttr: map[string]string{
				"authorized": "yes",
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			var mod setAttributes
			err := json.Unmarshal([]byte(tC.module), &mod)
			if err != nil {
				t.Fatalf("unexpected error unmarshalling module: %v", err)
			}
			state := tC.state
			if state == nil {
				state = testCallState{}.init()
			}
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
				t.Errorf("expected next of '%s' but got '%s'", tC.exp, nextStr)
			}
			if tC.expAttr != nil && !reflect.DeepEqual(state.contactData, tC.expAttr) {
				t.Errorf("expected contact data of '%s' but got '%s'", tC.expAttr, state.contactData)
			}
		})
	}
}
