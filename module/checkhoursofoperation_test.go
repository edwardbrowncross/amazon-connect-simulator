package module

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

func TestCheckHoursOfOperation(t *testing.T) {
	jsonOK := `{
		"id":"43dcc4f2-3392-4a38-90ed-0216f8594ea8",
		"type":"CheckHoursOfOperation",
		"branches":[
			{"condition":"True", "transition":"00000000-0000-4000-0000-000000000001"},
			{"condition":"False","transition":"00000000-0000-4000-0000-000000000002"},
			{"condition":"Error","transition":"00000000-0000-4000-0000-000000000003"}
		],
		"parameters":[]
	}`
	jsonOKCustomHours := `{
		"id":"43dcc4f2-3392-4a38-90ed-0216f8594ea8",
		"type":"CheckHoursOfOperation",
		"branches":[
			{"condition":"True", "transition":"00000000-0000-4000-0000-000000000001"},
			{"condition":"False","transition":"00000000-0000-4000-0000-000000000002"},
			{"condition":"Error","transition":"00000000-0000-4000-0000-000000000003"}
		],
		"parameters":[
			{ "name": "Hours", "value": "arn:aws:connect:eu-west-2:456789012345:instance/ffffffff-ffff-4000-ffff-ffffffffffff/contact-flow/8f135e8d-278d-4c7a-9415-3b3b79a5d07c", "resourceName": "Custom Hours" }
		]
	}`
	jsonBadMod := `{
		"id":"43dcc4f2-3392-4a38-90ed-0216f8594ea8",
		"type":"Disconnect"
	}`
	testCases := []struct {
		desc   string
		module string
		state  *testCallState
		exp    string
		expErr string
	}{
		{
			desc:   "wrong module",
			module: jsonBadMod,
			expErr: "module of type Disconnect being run as checkHoursOfOperation",
		},
		{
			desc:   "Current Queue not set",
			module: jsonOK,
			state: testCallState{
				inHours: func(n string, q bool, ct time.Time) (bool, error) {
					t.Error("Expected inHours not to be called with no current queue set, but it was called")
					return true, nil
				},
			}.init(),
			exp: "00000000-0000-4000-0000-000000000003",
		},
		{
			desc:   "inHours returns error",
			module: jsonOKCustomHours,
			state: testCallState{
				inHours: func(n string, q bool, ct time.Time) (bool, error) {
					return false, errors.New("Not known")
				},
			}.init(),
			exp: "00000000-0000-4000-0000-000000000003",
		},
		{
			desc:   "In hours - current Queue",
			module: jsonOK,
			state: testCallState{
				system: map[flow.SystemKey]string{
					flow.SystemQueueName: "TestQueue",
				},
				inHours: func(n string, q bool, ct time.Time) (bool, error) {
					if n != "TestQueue" {
						t.Errorf("Expected queue name of %s but got %s", "TestQueue", n)
					}
					if !q {
						t.Error("Expected isQueue to be true but it was false")
					}
					return true, nil
				},
			}.init(),
			exp: "00000000-0000-4000-0000-000000000001",
		},
		{
			desc:   "Out of hours - current Queue",
			module: jsonOK,
			state: testCallState{
				system: map[flow.SystemKey]string{
					flow.SystemQueueName: "TestQueue",
				},
				inHours: func(n string, q bool, ct time.Time) (bool, error) {
					return false, nil
				},
			}.init(),
			exp: "00000000-0000-4000-0000-000000000002",
		},
		{
			desc:   "In hours - custom hours",
			module: jsonOKCustomHours,
			state: testCallState{
				inHours: func(n string, q bool, ct time.Time) (bool, error) {
					if n != "Custom Hours" {
						t.Errorf("Expected queue name of %s but got %s", "TestQueue", n)
					}
					if q {
						t.Error("Expected isQueue to be false but it was true")
					}
					return true, nil
				},
			}.init(),
			exp: "00000000-0000-4000-0000-000000000001",
		},
		{
			desc:   "Out of hours - custom hours",
			module: jsonOKCustomHours,
			state: testCallState{
				inHours: func(n string, q bool, ct time.Time) (bool, error) {
					return false, nil
				},
			}.init(),
			exp: "00000000-0000-4000-0000-000000000002",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			var mod checkHoursOfOperation
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
		})
	}
}
