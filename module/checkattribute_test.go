package module

import (
	"encoding/json"
	"testing"

	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

func TestCheckAttribute(t *testing.T) {
	jsonNum := `{
		"id":"43dcc4f2-3392-4a38-90ed-0216f8594ea8",
		"type":"CheckAttribute",
		"branches":[
			{"condition":"Evaluate","conditionType":"LessThanOrEqualTo","conditionValue":"3","transition":"00000000-0000-4000-0000-000000000001"},
			{"condition":"NoMatch","transition":"00000000-0000-4000-0000-000000000002"}
		],
		"parameters":[
			{"name":"Attribute","value":"securityAttempts"},
			{"name":"Namespace","value":"External"}
		]
	}`
	jsonString := `{
		"id":"43dcc4f2-3392-4a38-90ed-0216f8594ea8",
		"type":"CheckAttribute",
		"branches":[
			{"condition":"Evaluate","conditionType":"Equals","conditionValue":"complaints","transition":"00000000-0000-4000-0000-000000000001"},
			{"condition":"NoMatch","transition":"00000000-0000-4000-0000-000000000002"}
		],
		"parameters":[
			{"name":"Attribute","value":"Queue name"},
			{"name":"Namespace","value":"System"}
		]
	}`
	jsonBadMod := `{
		"id":"43dcc4f2-3392-4a38-90ed-0216f8594ea8",
		"type":"Disconnect"
	}`
	jsonBadParams := `{
		"id":"43dcc4f2-3392-4a38-90ed-0216f8594ea8",
		"type":"CheckAttribute",
		"parameters":[]
	}`
	jsonBadCondition := `{
		"id":"43dcc4f2-3392-4a38-90ed-0216f8594ea8",
		"type":"CheckAttribute",
		"branches":[
			{"condition":"Evaluate","conditionType":"StartsWith","conditionValue":"n","transition":"00000000-0000-4000-0000-000000000001"},	
			{"condition":"NoMatch","transition":"00000000-0000-4000-0000-000000000002"}
		],
		"parameters":[
			{"name":"Attribute","value":"Queue name"},
			{"name":"Namespace","value":"System"}
		]
	}`
	testCases := []struct {
		desc    string
		module  string
		context CallContext
		exp     string
		expErr  string
	}{
		{
			desc:    "wrong module",
			module:  jsonBadMod,
			context: testContext{}.init(),
			exp:     "",
			expErr:  "module of type Disconnect being run as checkAttribute",
		},
		{
			desc:    "bad parameters",
			module:  jsonBadParams,
			context: testContext{}.init(),
			exp:     "",
			expErr:  "missing parameter Namespace",
		},
		{
			desc:    "unknown condition",
			module:  jsonBadCondition,
			context: testContext{}.init(),
			exp:     "",
			expErr:  "unhandled condition type: StartsWith",
		},
		{
			desc:   "numeric comparison match",
			module: jsonNum,
			context: testContext{
				external: map[string]string{
					"securityAttempts": "3",
				},
			}.init(),
			exp: "00000000-0000-4000-0000-000000000001",
		},
		{
			desc:   "numeric comparison no match",
			module: jsonNum,
			context: testContext{
				external: map[string]string{
					"securityAttempts": "10",
				},
			}.init(),
			exp: "00000000-0000-4000-0000-000000000002",
		},
		{
			desc:   "string comparison match",
			module: jsonString,
			context: testContext{
				system: map[string]string{
					flow.SystemQueueName: "complaints",
				},
			}.init(),
			exp: "00000000-0000-4000-0000-000000000001",
		},
		{
			desc:   "string comparison no match",
			module: jsonString,
			context: testContext{
				system: map[string]string{
					flow.SystemQueueName: "sales",
				},
			}.init(),
			exp: "00000000-0000-4000-0000-000000000002",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			var mod checkAttribute
			err := json.Unmarshal([]byte(tC.module), &mod)
			if err != nil {
				t.Fatalf("unexpected error unmarshalling module: %v", err)
			}
			next, err := mod.Run(tC.context)
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

func TestEvaluateConditions(t *testing.T) {
	testJSON := []byte(`
		[
			{
				"condition": "Evaluate",
				"conditionType": "Equals",
				"conditionValue": "Raisins",
				"transition": "00000000-0000-4000-0000-000000000007"
			},
			{
				"condition": "Evaluate",
				"conditionType": "GreaterThanOrEqualTo",
				"conditionValue": "7",
				"transition": "00000000-0000-4000-0000-000000000001"
			},
			{
				"condition": "Evaluate",
				"conditionType": "GreaterThan",
				"conditionValue": "5",
				"transition": "00000000-0000-4000-0000-000000000002"
			},
			{
				"condition": "Evaluate",
				"conditionType": "Equals",
				"conditionValue": "4",
				"transition": "00000000-0000-4000-0000-000000000003"
			},
			{
				"condition": "Evaluate",
				"conditionType": "LessThanOrEqualTo",
				"conditionValue": "1",
				"transition": "00000000-0000-4000-0000-000000000005"
			},
			{
				"condition": "Evaluate",
				"conditionType": "LessThan",
				"conditionValue": "3",
				"transition": "00000000-0000-4000-0000-000000000004"
			},
			{
				"condition": "NoMatch",
				"transition": "00000000-0000-4000-0000-000000000006"
			}
		]
	`)
	var testBranches flow.ModuleBranchList
	err := json.Unmarshal(testJSON, &testBranches)
	if err != nil {
		t.Fatalf("failed to generate test branches: %v", err)
	}

	testCases := []struct {
		v   string
		exp string
	}{
		{v: "0", exp: "00000000-0000-4000-0000-000000000005"},
		{v: "1", exp: "00000000-0000-4000-0000-000000000005"},
		{v: "2", exp: "00000000-0000-4000-0000-000000000004"},
		{v: "3", exp: "00000000-0000-4000-0000-000000000006"},
		{v: "4", exp: "00000000-0000-4000-0000-000000000003"},
		{v: "5", exp: "00000000-0000-4000-0000-000000000006"},
		{v: "6", exp: "00000000-0000-4000-0000-000000000002"},
		{v: "7", exp: "00000000-0000-4000-0000-000000000001"},
		{v: "8", exp: "00000000-0000-4000-0000-000000000001"},
		{v: "60", exp: "00000000-0000-4000-0000-000000000001"},
		{v: "Raisins", exp: "00000000-0000-4000-0000-000000000007"},
	}
	for _, tC := range testCases {
		t.Run("input of "+tC.v, func(t *testing.T) {
			res, err := evaluateConditions(testBranches, tC.v)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if res == nil {
				t.Errorf("expected branch of %s but got nil", tC.exp)
			} else if string(*res) != tC.exp {
				t.Errorf("expected branch of %s but got %v", tC.exp, *res)
			}
		})
	}
}
