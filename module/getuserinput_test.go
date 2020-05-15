package module

import (
	"encoding/json"
	"testing"
	"time"
)

func TestGetUserInput(t *testing.T) {
	jsonBadMod := `{
		"id":"43dcc4f2-3392-4a38-90ed-0216f8594ea8",
		"type":"Transfer"
	}`
	jsonBadParams := `{
		"id":"43dcc4f2-3392-4a38-90ed-0216f8594ea8",
		"type":"GetUserInput",
		"branches":[],
		"parameters":[
			{"name":"TextToSpeechType","value":"ssml"},
			{"name":"Timeout","value":"5"},
			{"name":"MaxDigits","value":"1"}
		]
	}`
	jsonOK := `{
		"id":"0a7af980-635b-4965-adbd-76594b8dffec",
		"type":"GetUserInput",
		"branches":[
			{"condition":"Evaluate","conditionType":"Equals","conditionValue":"1","transition":"00000000-0000-4000-0000-000000000001"},
			{"condition":"Evaluate","conditionType":"Equals","conditionValue":"2","transition":"00000000-0000-4000-0000-000000000002"},
			{"condition":"Timeout","transition":"00000000-0000-4000-0000-000000000003"},
			{"condition":"NoMatch","transition":"00000000-0000-4000-0000-000000000004"},
			{"condition":"Error","transition":"00000000-0000-4000-0000-000000000005"}
		],
		"parameters":[
			{"name":"Text","value":"prompt","namespace":"External"},
			{"name":"TextToSpeechType","value":"ssml"},
			{"name":"Timeout","value":"5"},
			{"name":"MaxDigits","value":"1"}
		],
		"target":"Digits"
	}`
	testCases := []struct {
		desc          string
		module        string
		entry         string
		ctx           *testContext
		exp           string
		expErr        string
		expPrompt     string
		expRcvTimeout time.Duration
		expRcvCount   int
	}{
		{
			desc:   "wrong module",
			module: jsonBadMod,
			expErr: "module of type Transfer being run as getUserInput",
		},
		{
			desc:   "missing parameter",
			module: jsonBadParams,
			expErr: "missing parameter Text",
		},
		{
			desc:   "parameter reference missing",
			module: jsonOK,
			exp:    "00000000-0000-4000-0000-000000000005",
			ctx: testContext{
				external: map[string]string{
					"speakme": "<speak>Enter a number</speak>",
				},
			}.init(),
			expPrompt: "",
		},
		{
			desc:   "matching entry",
			module: jsonOK,
			entry:  "2",
			exp:    "00000000-0000-4000-0000-000000000002",
			ctx: testContext{
				external: map[string]string{
					"prompt": "<speak>Enter a number</speak>",
				},
			}.init(),
			expPrompt:     "<speak>Enter a number</speak>",
			expRcvCount:   1,
			expRcvTimeout: 5 * time.Second,
		},
		{
			desc:   "mismatched entry",
			module: jsonOK,
			entry:  "3",
			exp:    "00000000-0000-4000-0000-000000000004",
			ctx: testContext{
				external: map[string]string{
					"prompt": "<speak>Enter a number</speak>",
				},
			}.init(),
			expPrompt:     "<speak>Enter a number</speak>",
			expRcvCount:   1,
			expRcvTimeout: 5 * time.Second,
		},
		{
			desc:   "timeout",
			module: jsonOK,
			exp:    "00000000-0000-4000-0000-000000000003",
			ctx: testContext{
				external: map[string]string{
					"prompt": "<speak>Enter a number</speak>",
				},
			}.init(),
			expPrompt:     "<speak>Enter a number</speak>",
			expRcvCount:   1,
			expRcvTimeout: 5 * time.Second,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			var mod getUserInput
			err := json.Unmarshal([]byte(tC.module), &mod)
			if err != nil {
				t.Fatalf("unexpected error unmarshalling module: %v", err)
			}
			ctx := tC.ctx
			if ctx == nil {
				ctx = testContext{}.init()
			}
			ctx.i = tC.entry
			next, err := mod.Run(ctx)
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
			if ctx.o != tC.expPrompt {
				t.Errorf("expected prompt of '%s' but got '%s'", tC.expPrompt, ctx.o)
			}
			if ctx.rcv.count != tC.expRcvCount {
				t.Errorf("expected receive count of %d but got %d", ctx.rcv.count, tC.expRcvCount)
			}
			if ctx.rcv.timeout != tC.expRcvTimeout {
				t.Errorf("expected receive timeout of %d but got %d", ctx.rcv.timeout, tC.expRcvTimeout)
			}
		})
	}
}
