package module

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/edwardbrowncross/amazon-connect-simulator/call"
	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

type lambdaPayload struct {
	Details lambdaPayloadDetails `json:"Details"`
	Name    string               `json:"Name"`
}

type lambdaPayloadDetails struct {
	ContactData lambdaPayloadContactData `json:"ContactData"`
	Parameters  json.RawMessage          `json:"Parameters"`
}

type lambdaPayloadContactData struct {
	Attributes        json.RawMessage              `json:"Attributes"`
	ContactID         string                       `json:"ContactId"`
	InitialContactID  string                       `json:"InitialContactId"`
	PreviousContactID string                       `json:"PreviousContactId"`
	Channel           string                       `json:"Channel"`
	InitiationMethod  string                       `json:"InitiationMethod"`
	CustomerEndpoint  lambdaPayloadContactEndpoint `json:"CustomerEndpoint"`
	SystemEndpoint    lambdaPayloadContactEndpoint `json:"SystemEndpoint"`
	InstanceARN       string                       `json:"InstanceARN"`
	Queue             interface{}                  `json:"Queue"`
}

type lambdaPayloadContactEndpoint struct {
	Address string `json:"Address"`
	Type    string `json:"Type"`
}

type lambdaPayloadQueue struct {
	Name string `json:"Name"`
	ARN  string `json:"ARN"`
}

type invokeExternalResource flow.Module

type invokeExternalResourceParams struct {
	TimeLimit   string
	FunctionArn string
	Parameter   []call.KeyValue
}

func (m invokeExternalResource) Run(ctx *call.Context) (next *flow.ModuleID, err error) {
	if m.Type != flow.ModuleInvokeExternalResource {
		return nil, fmt.Errorf("module of type %s being run as invokeExternalResource", m.Type)
	}
	p := invokeExternalResourceParams{}
	err = ctx.UnmarshalParameters(m.Parameters, &p)
	if err != nil {
		return
	}
	fn := ctx.GetLambda(p.FunctionArn)
	if fn == nil {
		return nil, fmt.Errorf("no function found for lambda %s", p.FunctionArn)
	}
	fields := make([]string, len(p.Parameter))
	for i, p := range p.Parameter {
		v, _ := json.Marshal(p.V)
		fields[i] = fmt.Sprintf(`"%s":%s`, p.K, v)
	}
	paramsIn := fmt.Sprintf(`{%s}`, strings.Join(fields, ","))
	payloadIn := lambdaPayload{
		Details: lambdaPayloadDetails{
			Parameters: json.RawMessage(paramsIn),
		},
	}
	jsonIn, _ := json.Marshal(payloadIn)
	jsonOut, err := m.invoke(fn, string(jsonIn))
	if err != nil {
		fmt.Println(paramsIn)
		fmt.Println("debug:" + err.Error())
		return m.Branches.GetLink(flow.BranchError), nil
	}
	out := map[string]interface{}{}
	err = json.Unmarshal([]byte(jsonOut), &out)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal json from lambda: %s", jsonOut)
	}
	ctx.External = map[string]string{}
	for k, v := range out {
		ctx.External[k] = fmt.Sprintf("%v", v)
	}
	return m.Branches.GetLink(flow.BranchSuccess), nil
}

func (m invokeExternalResource) invoke(fn interface{}, withJSON string) (outJSON string, err error) {
	fnv := reflect.ValueOf(fn)
	inputt := reflect.TypeOf(fn).In(1)
	in := reflect.New(inputt)
	err = json.Unmarshal([]byte(withJSON), in.Interface())
	if err != nil {
		return
	}
	response := fnv.Call([]reflect.Value{
		reflect.ValueOf(context.Background()),
		in.Elem(),
	})
	if errV, ok := response[1].Interface().(error); ok && errV != nil {
		return "", errV
	}
	out, err := json.Marshal(response[0].Interface())
	if err != nil {
		return
	}
	outJSON = string(out)
	return
}
