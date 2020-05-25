package simulator

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestValidateLambda(t *testing.T) {
	type testLambdaOutput struct{}
	testCases := []struct {
		desc string
		fn   interface{}
		exp  string
	}{
		{
			desc: "not a function",
			fn:   &struct{}{},
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
			err := validateLambda(tC.fn)
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

func TestInvokeLambda(t *testing.T) {
	validIn := `{
		"Details":{
			"ContactData":{
				"Attributes":{
					"userId":"edwardbrowncross",
					"authorized":"true"
				},
				"ContactId":"asdasd","InitialContactId":"fghfgh","PreviousContactId":"fghfgh","Channel":"","InitiationMethod":"",
				"CustomerEndpoint":{"Address":"","Type":""},"SystemEndpoint":{"Address":"","Type":""},
				"InstanceARN":"","Queue":""
			},
			"Parameters":{
				"requestType":"greeting"
			}
		},
		"Name":""
	}`

	testCases := []struct {
		desc      string
		in        string
		fn        interface{}
		expOut    string
		expOutErr string
		expErr    string
	}{
		{
			desc:   "invalid json",
			in:     `<xml />`,
			fn:     func(context.Context, LambdaPayload) (o struct{}, e error) { return },
			expErr: "invalid character '<' looking for beginning of value",
		},
		{
			desc: "invalid output",
			in:   validIn,
			fn: func(ctx context.Context, in LambdaPayload) (out struct{ Greet interface{} }, err error) {
				out.Greet = &out
				return
			},
			expErr: "json: unsupported value: encountered a cycle via *struct { Greet interface {} }",
		},
		{
			desc: "success - returns error",
			in:   validIn,
			fn: func(ctx context.Context, in LambdaPayload) (out struct{ Greet string }, err error) {
				err = errors.New("missing dependency: left-pad.js")
				return
			},
			expOutErr: "missing dependency: left-pad.js",
		},
		{
			desc: "success",
			in:   validIn,
			fn: func(ctx context.Context, in LambdaPayload) (out struct {
				Greet string `json:"greet"`
			}, err error) {
				var param struct {
					ReqType string `json:"requestType"`
				}
				var attr struct {
					UserID string `json:"userId"`
					Auth   string `json:"authorized"`
				}
				if err = json.Unmarshal(in.Details.Parameters, &param); err != nil {
					return
				}
				if err = json.Unmarshal(in.Details.ContactData.Attributes, &attr); err != nil {
					return
				}
				if param.ReqType != "greeting" {
					t.Errorf("missing parameters: %s", in.Details.Parameters)
				}
				if attr.Auth != "true" {
					t.Errorf("missing attributes: %s", in.Details.ContactData)
				}
				out.Greet = "hello"
				return
			},
			expOut: `{"greet":"hello"}`,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			out, outErr, err := invokeLambda(tC.fn, tC.in)
			var outErrStr, errStr string
			if outErr != nil {
				outErrStr = outErr.Error()
			}
			if err != nil {
				errStr = err.Error()
			}
			if out != tC.expOut {
				t.Errorf("expected output of '%s'. Got '%s'", tC.expOut, out)
			}
			if outErrStr != tC.expOutErr {
				t.Errorf("expected output error of '%s'. Got '%s'", tC.expOutErr, outErrStr)
			}
			if errStr != tC.expErr {
				t.Errorf("expected error of '%s'. Got '%s'", tC.expErr, errStr)
			}
		})
	}
}
