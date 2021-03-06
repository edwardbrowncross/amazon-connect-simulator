package module

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/edwardbrowncross/amazon-connect-simulator/event"
	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

func TestTransfer(t *testing.T) {
	jsonBadMod := `{
		"id":"43dcc4f2-3392-4a38-90ed-0216f8594ea8",
		"type":"Disconnect"
	}`
	jsonFlowBadTarget := `{
		"id":"43dcc4f2-3392-4a38-90ed-0216f8594ea8",
		"type":"Transfer",
		"parameters":[],
		"target": "Skype"
	}`
	jsonFlowBadParam := `{
		"id":"43dcc4f2-3392-4a38-90ed-0216f8594ea8",
		"type":"Transfer",
		"parameters":[],
		"target": "Flow"
	}`
	jsonFlowOK := `{
		"id":"55c7b51c-ab55-4c63-ac42-235b4a0f904f",
		"type":"Transfer",
		"branches":[
			{"condition":"Error","transition":"00000000-0000-4000-0000-000000000002"}
		],
		"parameters":[{
			"name":"ContactFlowId",
			"value":"arn:aws:connect:eu-west-2:456789012345:instance/ffffffff-ffff-4000-ffff-ffffffffffff/contact-flow/ffffffff-0000-4000-0000-ffffffff0001",
			"resourceName":"Security"
		}],
		"target": "Flow"
	}`
	jsonQueueOK := `{
		"id":"55c7b51c-ab55-4c63-ac42-235b4a0f904f",
		"type":"Transfer",
		"branches":[
			{"condition":"AtCapacity","transition":"00000000-0000-4000-0000-000000000001"},
			{"condition":"Error","transition":"00000000-0000-4000-0000-000000000002"}
		],
		"parameters":[],
		"target": "Queue"
	}`
	jsonBlindNumberOK := `{
		"id":"55c7b51c-ab55-4c63-ac42-235b4a0f904f",
		"type":"Transfer",
		"branches":[
			{"condition":"Error","transition":"00000000-0000-4000-0000-000000000002"}
		],
		"parameters":[
			{"name":"TimeLimit","value":"30"},
			{"name":"BlindTransfer","value":true},
			{"name":"PhoneNumber","value":"+441234567890"}
		],
		"target": "PhoneNumber"
	}`
	jsonNumberOK := `{
		"id":"55c7b51c-ab55-4c63-ac42-235b4a0f904f",
		"type":"Transfer",
		"branches":[
			{"condition":"Success","transition":"00000000-0000-4000-0000-000000000001"},
			{"condition":"CallFailure","transition":"00000000-0000-4000-0000-000000000002"},
			{"condition":"Timeout","transition":"00000000-0000-4000-0000-000000000003"},
			{"condition":"Error","transition":"00000000-0000-4000-0000-000000000004"}
		],
		"parameters":[
			{"name":"TimeLimit","value":"30"},
			{"name":"BlindTransfer","value":false},
			{"name":"PhoneNumber","value":"+441234567890"}
		],
		"target": "PhoneNumber"
	}`
	testCases := []struct {
		desc   string
		module string
		state  *testCallState
		exp    string
		expSys map[flow.SystemKey]string
		expEvt []event.Event
		expErr string
	}{
		{
			desc:   "wrong module",
			module: jsonBadMod,
			expErr: "module of type Disconnect being run as transfer",
		},
		{
			desc:   "bad target",
			module: jsonFlowBadTarget,
			expErr: "unhandled transfer target: Skype",
		},
		{
			desc:   "bad parameter - flow",
			module: jsonFlowBadParam,
			expErr: "missing ContactFlowId parameter",
		},
		{
			desc:   "non-existant flow",
			module: jsonFlowOK,
			state: testCallState{
				flowStart: map[string]flow.ModuleID{},
			}.init(),
			exp:    "00000000-0000-4000-0000-000000000002",
			expEvt: []event.Event{},
		},
		{
			desc:   "no queue set",
			module: jsonQueueOK,
			state: testCallState{
				system: map[flow.SystemKey]string{},
			}.init(),
			exp:    "00000000-0000-4000-0000-000000000002",
			expEvt: []event.Event{},
		},
		{
			desc:   "success - flow",
			module: jsonFlowOK,
			state: testCallState{
				flowStart: map[string]flow.ModuleID{
					"Security": "00000000-0000-4000-0000-000000000001",
				},
			}.init(),
			exp: "00000000-0000-4000-0000-000000000001",
			expEvt: []event.Event{
				event.FlowTransferEvent{FlowName: "Security", FlowARN: "arn:aws:connect:eu-west-2:456789012345:instance/ffffffff-ffff-4000-ffff-ffffffffffff/contact-flow/ffffffff-0000-4000-0000-ffffffff0001"},
			},
		},
		{
			desc:   "success - queue",
			module: jsonQueueOK,
			state: testCallState{
				system: map[flow.SystemKey]string{
					flow.SystemQueueName: "complaints",
					flow.SystemQueueARN:  "arn:aws:connect:eu-west-2:456789012345:instance/ffffffff-ffff-4000-ffff-ffffffffffff/queue/ffffffff-0000-4000-0000-ffffffff0001",
				},
			}.init(),
			exp: "",
			expEvt: []event.Event{
				event.QueueTransferEvent{QueueName: "complaints", QueueARN: "arn:aws:connect:eu-west-2:456789012345:instance/ffffffff-ffff-4000-ffff-ffffffffffff/queue/ffffffff-0000-4000-0000-ffffffff0001"},
			},
		},
		{
			desc:   "success - blind number",
			module: jsonBlindNumberOK,
			exp:    "",
			expEvt: []event.Event{
				event.NumberTransferEvent{Tel: "+441234567890"},
			},
		},
		{
			desc:   "success - resumable number",
			module: jsonNumberOK,
			exp:    "00000000-0000-4000-0000-000000000001",
			expEvt: []event.Event{
				event.NumberTransferEvent{Tel: "+441234567890"},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			var mod transfer
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
