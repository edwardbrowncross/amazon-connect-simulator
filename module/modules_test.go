package module

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

type testCallState struct {
	i   string
	o   string
	rcv struct {
		count   int
		timeout time.Duration
	}
	external    map[string]string
	contactData map[string]string
	system      map[string]string
	lambda      map[string]interface{}
	flowStart   map[string]flow.ModuleID
}

func (st testCallState) init() *testCallState {
	if st.external == nil {
		st.external = map[string]string{}
	}
	if st.contactData == nil {
		st.contactData = map[string]string{}
	}
	if st.system == nil {
		st.system = map[string]string{}
	}
	if st.lambda == nil {
		st.lambda = map[string]interface{}{}
	}
	if st.flowStart == nil {
		st.flowStart = map[string]flow.ModuleID{}
	}
	return &st
}

func (st *testCallState) Send(s string) {
	st.o = s
}
func (st *testCallState) Receive(count int, timeout time.Duration) *string {
	st.rcv.count = count
	st.rcv.timeout = timeout
	if st.i == "" {
		return nil
	}
	return &st.i
}
func (st *testCallState) GetExternal(key string) interface{} {
	val, found := st.external[key]
	if !found {
		return nil
	}
	return val
}
func (st *testCallState) SetExternal(key string, value interface{}) {
	st.external[key] = fmt.Sprintf("%v", value)
}
func (st *testCallState) ClearExternal() {
	st.external = map[string]string{}
}
func (st *testCallState) GetContactData(key string) interface{} {
	val, found := st.contactData[key]
	if !found {
		return nil
	}
	return val
}
func (st *testCallState) SetContactData(key string, value interface{}) {
	st.contactData[key] = fmt.Sprintf("%v", value)
}
func (st *testCallState) GetSystem(key string) interface{} {
	val, found := st.system[key]
	if !found {
		return nil
	}
	return val
}
func (st *testCallState) SetSystem(key string, value interface{}) {
	st.system[key] = fmt.Sprintf("%v", value)
}
func (st *testCallState) GetLambda(named string) interface{} {
	return st.lambda[named]
}
func (st *testCallState) GetFlowStart(flowName string) *flow.ModuleID {
	r := st.flowStart[flowName]
	return &r
}

func TestMakeRunner(t *testing.T) {
	testCases := []struct {
		desc   string
		module string
		exp    Runner
	}{
		{
			desc:   "StoreUserInput",
			module: `{ "type": "StoreUserInput" }`,
			exp:    storeUserInput{},
		},
		{
			desc:   "CheckAttribute",
			module: `{ "type": "CheckAttribute" }`,
			exp:    checkAttribute{},
		},
		{
			desc:   "Transfer",
			module: `{ "type": "Transfer" }`,
			exp:    transfer{},
		},
		{
			desc:   "PlayPrompt",
			module: `{ "type": "PlayPrompt" }`,
			exp:    playPrompt{},
		},
		{
			desc:   "Disconnect",
			module: `{ "type": "Disconnect" }`,
			exp:    disconnect{},
		},
		{
			desc:   "SetQueue",
			module: `{ "type": "SetQueue" }`,
			exp:    setQueue{},
		},
		{
			desc:   "GetUserInput",
			module: `{ "type": "GetUserInput" }`,
			exp:    getUserInput{},
		},
		{
			desc:   "SetAttributes",
			module: `{ "type": "SetAttributes" }`,
			exp:    setAttributes{},
		},
		{
			desc:   "InvokeExternalResource",
			module: `{ "type": "InvokeExternalResource" }`,
			exp:    invokeExternalResource{},
		},
		{
			desc:   "CheckHoursOfOperation",
			module: `{ "type": "CheckHoursOfOperation" }`,
			exp:    checkHoursOfOperation{},
		},
		{
			desc:   "Passthrough",
			module: `{ "type": "WhatIsThisIDontEven" }`,
			exp:    passthrough{},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			m := flow.Module{}
			err := json.Unmarshal([]byte(tC.module), &m)
			if err != nil {
				t.Fatalf("unexpected error parsing module: %v", err)
			}
			r := MakeRunner(m)
			if reflect.TypeOf(r) != reflect.TypeOf(tC.exp) {
				t.Errorf("expected type of %v but got %v", reflect.TypeOf(tC.exp), reflect.TypeOf(r))
			}
		})
	}
}

func TestUnmarshalErrors(t *testing.T) {
	testCases := []struct {
		desc   string
		params string
		into   interface{}
		expErr string
	}{
		{
			desc:   "not a pointer",
			params: "[]",
			into:   struct{}{},
			expErr: "second parameter should be non-nil pointer",
		},
		{
			desc:   "nil pointer",
			params: "[]",
			into:   (*struct{})(nil),
			expErr: "second parameter should be non-nil pointer",
		},
		{
			desc:   "single bad namespace",
			params: `[{"name":"File","value":"bucket","namespace":"S3"}]`,
			into:   &struct{ File string }{},
			expErr: "unknown namespace: S3",
		},
		{
			desc:   "single missing parameter",
			params: `[{"name":"File","value":"index.html"}]`,
			into:   &struct{ Directory string }{},
			expErr: "missing parameter Directory",
		},
		{
			desc:   "single type mismatch",
			params: `[{"name":"File","value":"index.html"}]`,
			into:   &struct{ File int }{},
			expErr: "type mismatch in field File. Cannot convert string to int",
		},
		{
			desc: "slice bad namespace",
			params: `[
				{"name":"File","value":"bucket","namespace":"S3"},
				{"name":"File","value":"index.html"}
			]`,
			into:   &struct{ File []string }{},
			expErr: "unknown namespace: S3",
		},
		{
			desc: "slice type mismatch",
			params: `[
				{"name":"Value","value":"5"},
				{"name":"Value","value":10}
			]`,
			into:   &struct{ Value []string }{},
			expErr: "type mismatch in field Value. Cannot convert float64 to string",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			pl := flow.ModuleParameterList{}
			if err := json.Unmarshal([]byte(tC.params), &pl); err != nil {
				t.Fatalf("unexpected error unmarshalling parameters: %v", err)
			}
			pr := parameterResolver{testCallState{}.init()}
			err := pr.unmarshal(pl, tC.into)
			errStr := ""
			if err != nil {
				errStr = err.Error()
			}
			if errStr != tC.expErr {
				t.Errorf("expected error of '%v' but got '%v'", tC.expErr, errStr)
			}
		})
	}
}

func TestUnmarshalOK(t *testing.T) {
	testJSON := `{
		"parameters":[
			{"name":"Text","value":"Please enter your date of birth.","namespace":null},
			{"name":"TextToSpeechType","value":"text"},
			{"name":"Timeout","value":"5"},
			{"name":"MaxDigits","value":8},
			{"name":"EncryptEntry","value":true},
			{"name":"Parameter","value":"testValue","key":"testKey1"},
			{"name":"Parameter","value":"testValue2","key":"testKey2","namespace":"External"},
			{"name":"Parameter","value":"testValue3","key":"testKey3","namespace":"System"},
			{"name":"Parameter","value":"testValue4","key":"testKey4","namespace":"User Defined"},
			{"name":"MissingVal","value":"missingKey","namespace":"User Defined"},
			{"name":"MissingSliceVal","value":"missingKey","namespace":"User Defined"}
		]
	}`
	m := flow.Module{}
	err := json.Unmarshal([]byte(testJSON), &m)
	if err != nil {
		t.Fatalf("error perparing parameters: %v", err)
	}
	c := testCallState{
		external: map[string]string{
			"testValue2": "foo",
		},
		system: map[string]string{
			"testValue3": "bar",
		},
		contactData: map[string]string{
			"testValue4": "baz",
		},
	}.init()

	type someText string
	into := struct {
		Text            someText
		Timeout         string
		MaxDigits       int
		EncryptEntry    bool
		Parameter       []flow.KeyValue
		MissingVal      int
		MissingSlice    []int
		MissingSliceVal []int
	}{}

	pr := parameterResolver{c}
	err = pr.unmarshal(m.Parameters, &into)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if into.Text != "Please enter your date of birth." {
		t.Errorf("expected Text of %v but got %v", "Please enter your date of birth.", into.Text)
	}
	if into.Timeout != "5" {
		t.Errorf("expected Timeout of %v but got %v", "5", into.Timeout)
	}
	if into.MaxDigits != 8 {
		t.Errorf("expected MaxDigits of %v but got %v", 8, into.MaxDigits)
	}
	if into.EncryptEntry != true {
		t.Errorf("expected EncryptEntry of %v but got %v", true, into.EncryptEntry)
	}
	if into.MissingVal != 0 {
		t.Errorf("expected MissingVal of %v but got %v", 0, into.MissingVal)
	}
	if len(into.MissingSlice) != 0 {
		t.Errorf("expected MissingSlice of %v but got %v", []int{}, into.MissingSlice)
	}
	expParam := []flow.KeyValue{
		{K: "testKey1", V: "testValue"},
		{K: "testKey2", V: "foo"},
		{K: "testKey3", V: "bar"},
		{K: "testKey4", V: "baz"},
	}
	if !reflect.DeepEqual(into.Parameter, expParam) {
		t.Errorf("expected Parameters of %v but got %v", expParam, into.Parameter)
	}
}
