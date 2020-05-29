package module

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/edwardbrowncross/amazon-connect-simulator/event"
	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

func TestInvokeExternalResource(t *testing.T) {
	jsonBadMod := `{
		"id":"43dcc4f2-3392-4a38-90ed-0216f8594ea8",
		"type":"Transfer"
	}`
	jsonBadParam := `{
		"id": "38cd099e-e9f0-4af2-ac6a-186fa89c6d1e",
		"type": "InvokeExternalResource",
		"branches": [
			{"condition":"Success","transition":"00000000-0000-4000-0000-000000000001"},
			{"condition":"Error","transition":"00000000-0000-4000-0000-000000000002"}
		],
		"parameters":[
			{"name":"TimeLimit","value":"3"}
		],
		"target": "Lambda"
	}`
	jsonBadTarget := `{
		"id": "38cd099e-e9f0-4af2-ac6a-186fa89c6d1e",
		"type": "InvokeExternalResource",
		"branches": [
			{"condition":"Success","transition":"00000000-0000-4000-0000-000000000001"},
			{"condition":"Error","transition":"00000000-0000-4000-0000-000000000002"}
		],
		"parameters":[
			{"name":"FunctionArn","value":"arn:aws:lambda:eu-west-2:456789012345:function:my-lambda-fn","namespace":null},
			{"name":"TimeLimit","value":"3"}
		],
		"target": "EC2"
	}`
	jsonOKNoParams := `{
		"id": "38cd099e-e9f0-4af2-ac6a-186fa89c6d1e",
		"type": "InvokeExternalResource",
		"branches": [
			{"condition":"Success","transition":"00000000-0000-4000-0000-000000000001"},
			{"condition":"Error","transition":"00000000-0000-4000-0000-000000000002"}
		],
		"parameters":[
			{"name":"FunctionArn","value":"arn:aws:lambda:eu-west-2:456789012345:function:my-lambda-fn","namespace":null},
			{"name":"TimeLimit","value":"1"}
		],
		"target": "Lambda"
	}`
	jsonOK := `{
		"id": "38cd099e-e9f0-4af2-ac6a-186fa89c6d1e",
		"type": "InvokeExternalResource",
		"branches": [
			{"condition":"Success","transition":"00000000-0000-4000-0000-000000000001"},
			{"condition":"Error","transition":"00000000-0000-4000-0000-000000000002"}
		],
		"parameters":[
			{"name":"FunctionArn","value":"arn:aws:lambda:eu-west-2:456789012345:function:my-lambda-fn","namespace":null},
			{"name":"TimeLimit","value":"1"},
			{"name":"Parameter","key":"input","value":"Stored customer input","namespace":"System"},
			{"name":"Parameter","key":"prevCount","value":"count","namespace":"External"}
		],
		"target": "Lambda"
	}`
	type testLambdaInput struct {
		C string `json:"prevCount"`
		I string `json:"input"`
	}
	type testLambdaOutput struct {
		V string `json:"newCount"`
	}
	testCases := []struct {
		desc           string
		module         string
		state          *testCallState
		exp            string
		expExternal    map[string]string
		expEvt         []event.Event
		expLambdaName  string
		expLambdaInput string
		expErr         string
	}{
		{
			desc:   "wrong module",
			module: jsonBadMod,
			expErr: "module of type Transfer being run as invokeExternalResource",
		},
		{
			desc:   "missing parameter",
			module: jsonBadParam,
			expErr: "missing parameter FunctionArn",
		},
		{
			desc:   "bad target",
			module: jsonBadTarget,
			expErr: "unknown target: EC2",
		},
		{
			desc:   "missing lambda",
			module: jsonOKNoParams,
			state: testCallState{
				lambdaErr: errors.New("lambda not found"),
			}.init(),
			expLambdaName:  "arn:aws:lambda:eu-west-2:456789012345:function:my-lambda-fn",
			expLambdaInput: `{}`,
			exp:            "00000000-0000-4000-0000-000000000002",
			expEvt:         []event.Event{},
			expErr:         "",
		},
		{
			desc:   "lambda error",
			module: jsonOKNoParams,
			state: testCallState{
				lambdaOutErr: errors.New("failed to do a thing"),
			}.init(),
			expLambdaName:  "arn:aws:lambda:eu-west-2:456789012345:function:my-lambda-fn",
			expLambdaInput: `{}`,
			exp:            "00000000-0000-4000-0000-000000000002",
			expEvt:         []event.Event{},
			expErr:         "",
		},
		{
			desc:   "json unmarshal error (not anticipated)",
			module: jsonOKNoParams,
			state: testCallState{
				lambdaOut: `<xml />`,
			}.init(),
			expLambdaName:  "arn:aws:lambda:eu-west-2:456789012345:function:my-lambda-fn",
			expLambdaInput: `{}`,
			expEvt:         []event.Event{},
			expErr:         "failed to unmarshal json from lambda: <xml />",
		},
		{
			desc:   "success",
			module: jsonOK,
			state: testCallState{
				system: map[flow.SystemKey]string{
					flow.SystemLastUserInput: "12345",
				},
				external: map[string]string{
					"count": "4",
				},
				lambdaOut: `{"newCount": 5}`,
			}.init(),
			expLambdaName:  "arn:aws:lambda:eu-west-2:456789012345:function:my-lambda-fn",
			expLambdaInput: `{"input":"12345","prevCount":"4"}`,
			exp:            "00000000-0000-4000-0000-000000000001",
			expEvt:         []event.Event{},
			expExternal: map[string]string{
				"newCount": "5",
			},
			expErr: "",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			var mod invokeExternalResource
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
				t.Errorf("expected next of '%s' but got '%v'", tC.exp, *next)
			}
			if tC.expLambdaName != state.lambdaIn.name {
				t.Errorf("expected to call lambda with ARN '%s' but called '%s'", tC.expLambdaName, state.lambdaIn.name)
			}
			if tC.expLambdaInput != string(state.lambdaIn.input) {
				t.Errorf("expected to call lambda with input of '%s' but called '%s'", tC.expLambdaInput, state.lambdaIn.input)
			}
			if tC.expExternal != nil && !reflect.DeepEqual(tC.expExternal, state.external) {
				t.Errorf("expected external to be:\n%v\nbut it was \n%v", tC.expExternal, state.external)
			}
			if (tC.expEvt != nil && !reflect.DeepEqual(tC.expEvt, state.events)) || (tC.expEvt == nil && len(state.events) > 0) {
				t.Errorf("expected events of '%v' but got '%v'", tC.expEvt, state.events)
			}
		})
	}
}
