package module

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/edwardbrowncross/amazon-connect-simulator/event"
)

func TestDisconnect(t *testing.T) {
	jsonBadMod := `{
		"id":"43dcc4f2-3392-4a38-90ed-0216f8594ea8",
		"type":"Transfer"
	}`
	jsonOK := `{
		"id":"55c7b51c-ab55-4c63-ac42-235b4a0f904f",
		"type":"Disconnect",
		"branches":[],
		"parameters":[]
	}`
	testCases := []struct {
		desc   string
		module string
		exp    string
		expEvt []event.Event
		expErr string
	}{
		{
			desc:   "wrong module",
			module: jsonBadMod,
			expErr: "module of type Transfer being run as disconnect",
		},
		{
			desc:   "success",
			module: jsonOK,
			exp:    "",
			expEvt: []event.Event{
				event.ModuleEvent{ID: "55c7b51c-ab55-4c63-ac42-235b4a0f904f", ModuleType: "Disconnect"},
				event.DisconnectEvent{},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			var mod disconnect
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
			if (tC.expEvt != nil && !reflect.DeepEqual(tC.expEvt, state.events)) || (tC.expEvt == nil && len(state.events) > 0) {
				t.Errorf("expected events of '%v' but got '%v'", tC.expEvt, state.events)
			}
		})
	}
}
