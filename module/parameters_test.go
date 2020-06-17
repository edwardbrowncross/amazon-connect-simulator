package module

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

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
		{
			desc: "pointer type mismatch",
			params: `[
				{"name":"Value","value":10}
			]`,
			into:   &struct{ Value *string }{},
			expErr: "type mismatch in field Value. Cannot convert float64 to string",
		},
		{
			desc: "pointer bad namespace",
			params: `[
				{"name":"File","value":"bucket","namespace":"S3"}
			]`,
			into:   &struct{ File *string }{},
			expErr: "unknown namespace: S3",
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
			{"name":"MissingSliceVal","value":"missingKey","namespace":"User Defined"},
			{"name":"PresentOptionalValue","value":"I am here"}
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
		system: map[flow.SystemKey]string{
			"testValue3": "bar",
		},
		contactData: map[string]string{
			"testValue4": "baz",
		},
	}.init()

	type someText string
	into := struct {
		Text                 someText
		Timeout              string
		MaxDigits            int
		EncryptEntry         bool
		Parameter            []flow.KeyValue
		MissingVal           int
		MissingSlice         []int
		MissingSliceVal      []int
		OptionalValue        *string
		PresentOptionalValue *string
	}{}

	pr := parameterResolver{c}
	err = pr.unmarshal(m.Parameters, &into)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if into.Text != "Please enter your date of birth." {
		t.Errorf("expected Text of %s but got %s", "Please enter your date of birth.", into.Text)
	}
	if into.Timeout != "5" {
		t.Errorf("expected Timeout of %s but got %s", "5", into.Timeout)
	}
	if into.MaxDigits != 8 {
		t.Errorf("expected MaxDigits of %d but got %d", 8, into.MaxDigits)
	}
	if into.EncryptEntry != true {
		t.Errorf("expected EncryptEntry of %v but got %v", true, into.EncryptEntry)
	}
	if into.MissingVal != 0 {
		t.Errorf("expected MissingVal of %d but got %d", 0, into.MissingVal)
	}
	if len(into.MissingSlice) != 0 {
		t.Errorf("expected MissingSlice of %v but got %v", []int{}, into.MissingSlice)
	}
	if into.PresentOptionalValue == nil || *into.PresentOptionalValue != "I am here" {
		t.Errorf("expected PresentOptionalValue of %v but got %v", "I am here", *into.PresentOptionalValue)
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
