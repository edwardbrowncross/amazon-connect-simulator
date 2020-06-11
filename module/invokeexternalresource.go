package module

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

type invokeExternalResource flow.Module

type invokeExternalResourceParams struct {
	TimeLimit   string
	FunctionArn string
	Parameter   []flow.KeyValue
}

func (m invokeExternalResource) Run(call CallConnector) (next *flow.ModuleID, err error) {
	if m.Type != flow.ModuleInvokeExternalResource {
		return nil, fmt.Errorf("module of type %s being run as invokeExternalResource", m.Type)
	}
	if m.Target != flow.TargetLambda {
		return nil, fmt.Errorf("unknown target: %s", m.Target)
	}
	p := invokeExternalResourceParams{}
	err = parameterResolver{call}.unmarshal(m.Parameters, &p)
	if err != nil {
		return
	}
	tl, err := strconv.Atoi(p.TimeLimit)
	if err != nil {
		return nil, fmt.Errorf("invalid TimeLimit: %s", p.TimeLimit)
	}
	fields := make([]string, len(p.Parameter))
	for i, p := range p.Parameter {
		v, _ := json.Marshal(p.V)
		fields[i] = fmt.Sprintf(`"%s":%s`, p.K, v)
	}
	paramsIn := fmt.Sprintf(`{%s}`, strings.Join(fields, ","))
	jsonOut, errOut, err := call.InvokeLambda(p.FunctionArn, json.RawMessage(paramsIn), time.Duration(tl)*time.Second)
	if err != nil {
		return m.Branches.GetLink(flow.BranchError), nil
	}
	if errOut != nil {
		return m.Branches.GetLink(flow.BranchError), nil
	}
	out := map[string]interface{}{}
	err = json.Unmarshal([]byte(jsonOut), &out)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal json from lambda: %s", jsonOut)
	}
	call.ClearExternal()
	for k, v := range out {
		call.SetExternal(k, v)
	}
	return m.Branches.GetLink(flow.BranchSuccess), nil
}
