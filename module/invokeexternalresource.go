package module

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

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
	Parameter   []flow.KeyValue
}

func (m invokeExternalResource) Run(ctx CallContext) (next *flow.ModuleID, err error) {
	if m.Type != flow.ModuleInvokeExternalResource {
		return nil, fmt.Errorf("module of type %s being run as invokeExternalResource", m.Type)
	}
	if m.Target != flow.TargetLambda {
		return nil, fmt.Errorf("unknown target: %s", m.Target)
	}
	p := invokeExternalResourceParams{}
	err = parameterResolver{ctx}.unmarshal(m.Parameters, &p)
	if err != nil {
		return
	}
	fn := ctx.GetLambda(p.FunctionArn)
	if fn == nil {
		return m.Branches.GetLink(flow.BranchError), nil
	}
	if ValidateLambda(fn) != nil {
		return m.Branches.GetLink(flow.BranchError), nil
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
	ctx.ClearExternal()
	for k, v := range out {
		ctx.SetExternal(k, v)
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

// ValidateLambda checks that a function has the signature required for execution by an invokeExternalResource block.
func ValidateLambda(fn interface{}) error {
	fnt := reflect.TypeOf(fn)
	if fnt.Kind() != reflect.Func {
		return fmt.Errorf("wanted function but got %s", fnt.Kind())
	}
	if fnt.NumIn() != 2 {
		return fmt.Errorf("expected function to take 2 parameters but it takes %d", fnt.NumIn())
	}
	contextt := reflect.TypeOf((*context.Context)(nil)).Elem()
	if !fnt.In(0).Implements(contextt) {
		return errors.New("expected first argument to be a context.Context")
	}
	if fnt.In(1).Kind() != reflect.Struct {
		return fmt.Errorf("expected second argument to be struct but it was: %s", fnt.In(1).Kind())
	}
	if fnt.NumOut() != 2 {
		return fmt.Errorf("expected function to return 2 elements but it returns %d", fnt.NumOut())
	}
	if fnt.Out(0).Kind() != reflect.Struct {
		return fmt.Errorf("expected first return to be struct but it was: %s", fnt.Out(0).Kind())
	}
	errort := reflect.TypeOf((*error)(nil)).Elem()
	if !fnt.Out(1).Implements(errort) {
		return errors.New("expected second return to be an error")
	}
	return nil
}
