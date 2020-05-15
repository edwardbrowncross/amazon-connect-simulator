package module

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

func TestStoreUserInput(t *testing.T) {
	jsonBadMod := `{
		"id":"43dcc4f2-3392-4a38-90ed-0216f8594ea8",
		"type":"Transfer"
	}`
	jsonBadParam := `{
		"id":"43dcc4f2-3392-4a38-90ed-0216f8594ea8",
		"type":"StoreUserInput",
		"parameters":[]
	}`
	jsonBadTimeout := `{
		"id":"55c7b51c-ab55-4c63-ac42-235b4a0f904f",
		"type":"StoreUserInput",
		"branches":[],
		"parameters":[
			{"name":"Text","value":"prompt","namespace":"External"},
			{"name":"Timeout","value":"fishcake"},
			{"name":"MaxDigits","value":8}
		]
	}`
	jsonOK := `{
		"id":"55c7b51c-ab55-4c63-ac42-235b4a0f904f",
		"type":"StoreUserInput",
		"branches":[
			{"condition":"Success","transition":"00000000-0000-4000-0000-000000000001"},
			{"condition":"Error","transition":"00000000-0000-4000-0000-000000000002"}
		],
		"parameters":[
			{"name":"Text","value":"prompt","namespace":"External"},
			{"name":"TextToSpeechType","value":"text"},
			{"name":"CustomerInputType","value":"Custom"},
			{"name":"Timeout","value":"7"},
			{"name":"MaxDigits","value":8},
			{"name":"EncryptEntry","value":false},
			{"name":"DisableCancel","value":true}
		]
	}`
	testCases := []struct {
		desc          string
		module        string
		ctx           *testContext
		exp           string
		expPrompt     string
		expErr        string
		expSys        map[string]string
		expRcvTimeout time.Duration
		expRcvCount   int
	}{
		{
			desc:   "wrong module",
			module: jsonBadMod,
			expErr: "module of type Transfer being run as storeUserInput",
		},
		{
			desc:   "missing parameter",
			module: jsonBadParam,
			expErr: "missing parameter Text",
		},
		{
			desc:   "bad timeout parameter",
			module: jsonBadTimeout,
			expErr: `strconv.Atoi: parsing "fishcake": invalid syntax`,
		},
		{
			desc:   "timeout",
			module: jsonOK,
			ctx: testContext{
				i: "",
			}.init(),
			exp: "00000000-0000-4000-0000-000000000002",
		},
		{
			desc:   "success",
			module: jsonOK,
			exp:    "00000000-0000-4000-0000-000000000001",
			ctx: testContext{
				i: "12345678",
				external: map[string]string{
					"prompt": "Please enter your account number.",
				},
			}.init(),
			expSys: map[string]string{
				string(flow.SystemLastUserInput): "12345678",
			},
			expPrompt:     "Please enter your account number.",
			expRcvCount:   8,
			expRcvTimeout: 7 * time.Second,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			var mod storeUserInput
			err := json.Unmarshal([]byte(tC.module), &mod)
			if err != nil {
				t.Fatalf("unexpected error unmarshalling module: %v", err)
			}
			ctx := tC.ctx
			if ctx == nil {
				ctx = testContext{}.init()
			}
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
				t.Errorf("expected next of '%s' but got '%s'", tC.exp, nextStr)
			}
			for k, v := range tC.expSys {
				if ctx.system[k] != v {
					t.Errorf("expected system %s to be '%s' but it was '%s'", k, v, ctx.system[k])
				}
			}
			if ctx.o != tC.expPrompt {
				t.Errorf("expected prompt of '%s' but got '%s'", tC.expPrompt, ctx.o)
			}
			if tC.expRcvCount > 0 && ctx.rcv.count != tC.expRcvCount {
				t.Errorf("expected receive count of %d but got %d", ctx.rcv.count, tC.expRcvCount)
			}
			if tC.expRcvTimeout > 0 && ctx.rcv.timeout != tC.expRcvTimeout {
				t.Errorf("expected receive timeout of %d but got %d", ctx.rcv.timeout, tC.expRcvTimeout)
			}
		})
	}
}
