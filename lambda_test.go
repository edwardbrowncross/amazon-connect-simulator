package simulator

import (
	"context"
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
