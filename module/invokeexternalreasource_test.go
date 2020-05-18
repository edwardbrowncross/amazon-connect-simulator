package module

import (
	"context"
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
		desc        string
		module      string
		state       *testCallState
		exp         string
		expExternal map[string]string
		expEvt      []event.Event
		expErr      string
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
			module: jsonOK,
			state: testCallState{
				lambda: map[string]interface{}{
					"arn:aws:lambda:eu-west-2:456789012345:function:a-different-lambda": func() {},
				},
			}.init(),
			exp: "00000000-0000-4000-0000-000000000002",
			expEvt: []event.Event{
				event.ModuleEvent{ID: "38cd099e-e9f0-4af2-ac6a-186fa89c6d1e", ModuleType: "InvokeExternalResource"},
			},
			expErr: "",
		},
		{
			desc:   "bad lambda signature",
			module: jsonOKNoParams,
			state: testCallState{
				lambda: map[string]interface{}{
					"arn:aws:lambda:eu-west-2:456789012345:function:my-lambda-fn": func(string) error {
						return nil
					},
				},
			}.init(),
			exp: "00000000-0000-4000-0000-000000000002",
			expEvt: []event.Event{
				event.ModuleEvent{ID: "38cd099e-e9f0-4af2-ac6a-186fa89c6d1e", ModuleType: "InvokeExternalResource"},
			},
			expErr: "",
		},
		{
			desc:   "lambda error",
			module: jsonOKNoParams,
			state: testCallState{
				lambda: map[string]interface{}{
					"arn:aws:lambda:eu-west-2:456789012345:function:my-lambda-fn": func(c context.Context, evt LambdaPayload) (out testLambdaOutput, err error) {
						return out, errors.New("something went wrong")
					},
				},
			}.init(),
			exp: "00000000-0000-4000-0000-000000000002",
			expEvt: []event.Event{
				event.ModuleEvent{ID: "38cd099e-e9f0-4af2-ac6a-186fa89c6d1e", ModuleType: "InvokeExternalResource"},
			},
			expErr: "",
		},
		{
			desc:   "success",
			module: jsonOK,
			state: testCallState{
				system: map[string]string{
					string(flow.SystemLastUserInput): "12345",
				},
				external: map[string]string{
					"count": "4",
				},
				lambda: map[string]interface{}{
					"arn:aws:lambda:eu-west-2:456789012345:function:my-lambda-fn": func(c context.Context, evt LambdaPayload) (out testLambdaOutput, err error) {
						in := testLambdaInput{}
						err = json.Unmarshal(evt.Details.Parameters, &in)
						if err != nil {
							t.Errorf("unexpectedly failed to unmarshal input: %s", evt.Details.Parameters)
							return
						}
						if in.C != "4" {
							t.Errorf("expected input count of 4 but got %s", in.C)
						}
						if in.I != "12345" {
							t.Errorf("expected input digits of 12345 but got %s", in.I)
						}
						return testLambdaOutput{
							V: "5",
						}, nil
					},
				},
			}.init(),
			exp: "00000000-0000-4000-0000-000000000001",
			expEvt: []event.Event{
				event.ModuleEvent{ID: "38cd099e-e9f0-4af2-ac6a-186fa89c6d1e", ModuleType: "InvokeExternalResource"},
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
			if tC.expExternal != nil && !reflect.DeepEqual(tC.expExternal, state.external) {
				t.Errorf("expected external to be:\n%v\nbut it was \n%v", tC.expExternal, state.external)
			}
			if (tC.expEvt != nil && !reflect.DeepEqual(tC.expEvt, state.events)) || (tC.expEvt == nil && len(state.events) > 0) {
				t.Errorf("expected events of '%v' but got '%v'", tC.expEvt, state.events)
			}
		})
	}
}

func TestValidateLambda(t *testing.T) {
	type testLambdaOutput struct{}
	testCases := []struct {
		desc string
		fn   interface{}
		exp  string
	}{
		{
			desc: "not a function",
			fn:   testCallState{}.init(),
			exp:  "wanted function but got ptr",
		},
		{
			desc: "one input",
			fn:   func(LambdaPayload) (o testLambdaOutput, e error) { return },
			exp:  "expected function to take 2 parameters but it takes 1",
		},
		{
			desc: "first parameter not context",
			fn:   func(interface{}, LambdaPayload) (o testLambdaOutput, e error) { return },
			exp:  "expected first argument to be a context.Context",
		},
		{
			desc: "second parameter not struct",
			fn:   func(context.Context, string) (o testLambdaOutput, e error) { return },
			exp:  "expected second argument to be struct but it was: string",
		},
		{
			desc: "only one return",
			fn:   func(context.Context, LambdaPayload) (o testLambdaOutput) { return },
			exp:  "expected function to return 2 elements but it returns 1",
		},
		{
			desc: "first return not a struct",
			fn:   func(context.Context, LambdaPayload) (s string, e error) { return },
			exp:  "expected first return to be struct but it was: string",
		},
		{
			desc: "second return not an error",
			fn:   func(context.Context, LambdaPayload) (o testLambdaOutput, e string) { return },
			exp:  "expected second return to be an error",
		},
		{
			desc: "success",
			fn:   func(context.Context, LambdaPayload) (o testLambdaOutput, e error) { return },
			exp:  "",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			err := ValidateLambda(tC.fn)
			errStr := ""
			if err != nil {
				errStr = err.Error()
			}
			if errStr != tC.exp {
				t.Errorf("expected error of '%s' but got '%s'", tC.exp, errStr)
			}
		})
	}
}
