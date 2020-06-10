package simulator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

// LambdaPayload is the shape of an Amazon Connect lambda event.
type LambdaPayload struct {
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

// validateLambda checks that a function has the signature required for execution by an invokeExternalResource block.
func validateLambda(fn interface{}) error {
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
	// if fnt.Out(0).Kind() != reflect.Struct {
	// 	return fmt.Errorf("expected first return to be struct but it was: %s", fnt.Out(0).Kind())
	// }
	errort := reflect.TypeOf((*error)(nil)).Elem()
	if !fnt.Out(1).Implements(errort) {
		return errors.New("expected second return to be an error")
	}
	return nil
}

func invokeLambda(fn interface{}, inJSON string) (outJSON string, outErr error, err error) {
	fnv := reflect.ValueOf(fn)
	inputt := reflect.TypeOf(fn).In(1)
	in := reflect.New(inputt)
	err = json.Unmarshal([]byte(inJSON), in.Interface())
	if err != nil {
		return
	}
	response := fnv.Call([]reflect.Value{
		reflect.ValueOf(context.Background()),
		in.Elem(),
	})
	if outErr, ok := response[1].Interface().(error); ok && outErr != nil {
		return "", outErr, nil
	}
	out, err := json.Marshal(response[0].Interface())
	if err != nil {
		return
	}
	outJSON = string(out)
	return
}
