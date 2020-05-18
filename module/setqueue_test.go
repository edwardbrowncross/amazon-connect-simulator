package module

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/edwardbrowncross/amazon-connect-simulator/event"
	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

func TestSetQueue(t *testing.T) {
	jsonBadMod := `{
		"id":"43dcc4f2-3392-4a38-90ed-0216f8594ea8",
		"type":"Transfer"
	}`
	jsonBadParam := `{
		"id":"43dcc4f2-3392-4a38-90ed-0216f8594ea8",
		"type":"SetQueue",
		"parameters":[]
	}`
	jsonOK := `{
		"id":"55c7b51c-ab55-4c63-ac42-235b4a0f904f",
		"type":"SetQueue",
		"branches":[
			{"condition":"Success","transition":"00000000-0000-4000-0000-000000000001"},
			{"condition":"Error","transition":"00000000-0000-4000-0000-000000000002"}
		],
		"parameters":[{"name":"Queue","value":"arn:aws:connect:eu-west-2:456789012345:instance/ffffffff-ffff-4000-ffff-ffffffffffff/queue/ffffffff-0000-4000-0000-ffffffff0001","namespace":null,"resourceName":"Complaints"}]
	}`
	testCases := []struct {
		desc   string
		module string
		exp    string
		expSys map[string]string
		expEvt []event.Event
		expErr string
	}{
		{
			desc:   "wrong module",
			module: jsonBadMod,
			expErr: "module of type Transfer being run as setQueue",
		},
		{
			desc:   "bad parameter",
			module: jsonBadParam,
			expErr: "missing Queue parameter",
		},
		{
			desc:   "success",
			module: jsonOK,
			exp:    "00000000-0000-4000-0000-000000000001",
			expSys: map[string]string{
				string(flow.SystemQueueARN):  "arn:aws:connect:eu-west-2:456789012345:instance/ffffffff-ffff-4000-ffff-ffffffffffff/queue/ffffffff-0000-4000-0000-ffffffff0001",
				string(flow.SystemQueueName): "Complaints",
			},
			expEvt: []event.Event{
				event.ModuleEvent{ID: "55c7b51c-ab55-4c63-ac42-235b4a0f904f", ModuleType: "SetQueue"},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			var mod setQueue
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
				t.Errorf("expected next of '%s' but got '%s'", tC.exp, nextStr)
			}
			for k, v := range tC.expSys {
				if state.system[k] != v {
					t.Errorf("expected system %s to be '%s' but it was '%s'", k, v, state.system[k])
				}
			}
			if (tC.expEvt != nil && !reflect.DeepEqual(tC.expEvt, state.events)) || (tC.expEvt == nil && len(state.events) > 0) {
				t.Errorf("expected events of '%v' but got '%v'", tC.expEvt, state.events)
			}
		})
	}
}
