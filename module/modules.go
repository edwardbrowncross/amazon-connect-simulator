package module

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

// Runner takes a call context and returns the ID of the next block to run, or nil if the call is over.
type Runner interface {
	Run(CallContext) (*flow.ModuleID, error)
}

// MakeRunner takes the data of a module (block) and wraps it in a type that provides the functionality of that block.
func MakeRunner(m flow.Module) Runner {
	switch m.Type {
	case flow.ModuleStoreUserInput:
		return storeUserInput(m)
	case flow.ModuleCheckAttribute:
		return checkAttribute(m)
	case flow.ModuleTransfer:
		return transfer(m)
	case flow.ModulePlayPrompt:
		return playPrompt(m)
	case flow.ModuleDisconnect:
		return disconnect(m)
	case flow.ModuleSetQueue:
		return setQueue(m)
	case flow.ModuleGetUserInput:
		return getUserInput(m)
	case flow.ModuleSetAttributes:
		return setAttributes(m)
	case flow.ModuleInvokeExternalResource:
		return invokeExternalResource(m)
	case flow.ModuleCheckHoursOfOperation:
		return checkHoursOfOperation(m)
	default:
		return passthrough(m)
	}
}

// CallContext describes what a module needs to interact with the ongoing call.
type CallContext interface {
	Send(s string)
	Receive(count int, timeout time.Duration) *string
	GetExternal(key string) interface{}
	SetExternal(key string, value interface{})
	ClearExternal()
	GetContactData(key string) interface{}
	SetContactData(key string, value interface{})
	GetSystem(key string) interface{}
	SetSystem(key string, value interface{})
	GetLambda(named string) interface{}
	GetFlowStart(flowName string) *flow.ModuleID
}

type valueGetter interface {
	GetExternal(key string) interface{}
	GetContactData(key string) interface{}
	GetSystem(key string) interface{}
}

// parameterResolver uses the base methods of the CallContext to perform more sophisticated lookup operations.
type parameterResolver struct {
	valueGetter
}

// get gets a single value by namespace and key.
func (ctx parameterResolver) get(namespace flow.ModuleParameterNamespace, key string) (interface{}, error) {
	switch namespace {
	case flow.NamespaceUserDefined:
		return ctx.GetContactData(key), nil
	case flow.NamespaceExternal:
		return ctx.GetExternal(key), nil
	case flow.NamespaceSystem:
		return ctx.GetSystem(key), nil
	default:
		return nil, fmt.Errorf("unknown namespace: %s", namespace)
	}
}

// resolve takes a raw module parameter and looks up its value (whether static or dynamic).
func (ctx parameterResolver) resolve(p flow.ModuleParameter) (val interface{}, err error) {
	if p.Namespace == nil || *p.Namespace == "" {
		val = p.Value
	} else {
		key, ok := p.Value.(string)
		if !ok {
			return
		}
		val, err = ctx.get(*p.Namespace, key)
		if err != nil {
			return
		}
	}
	if p.Key != "" {
		val = flow.KeyValue{
			K: p.Key,
			V: fmt.Sprintf("%v", val),
		}
		return
	}
	return
}

// unmarshal takes the list of a block's parameters and unmarshals it into a typed struct.
// The struct field names should match the names of the parameters.
// Type checking will be performed. If the type of the value cannot be converted to the field type, it errors.
// Where there are multiple parameters with the same name, use a field with a slice type.
func (ctx parameterResolver) unmarshal(plist flow.ModuleParameterList, into interface{}) error {
	if reflect.ValueOf(into).Kind() != reflect.Ptr || into == nil || reflect.ValueOf(into).IsNil() {
		return errors.New("second parameter should be non-nil pointer")
	}
	intov := reflect.ValueOf(into).Elem()
	for i := 0; i < intov.NumField(); i++ {
		f := intov.Type().Field(i)
		switch f.Type.Kind() {
		case reflect.Slice:
			sliceType := f.Type.Elem()
			ps := plist.List(f.Name)
			vals := reflect.MakeSlice(reflect.SliceOf(sliceType), len(ps), len(ps))
			for j, p := range ps {
				val, err := ctx.resolve(p)
				if err != nil {
					return err
				}
				if val == nil {
					continue
				}
				if !reflect.TypeOf(val).ConvertibleTo(sliceType) {
					return fmt.Errorf("type mismatch in field %s. Cannot convert %s to %s", f.Name, reflect.TypeOf(val), sliceType)
				}
				vals.Index(j).Set(reflect.ValueOf(val).Convert(sliceType))
				intov.Field(i).Set(vals)
			}
		default:
			p := plist.Get(f.Name)
			if p == nil {
				return fmt.Errorf("missing parameter %s", f.Name)
			}
			val, err := ctx.resolve(*p)
			if err != nil {
				return err
			}
			if val == nil {
				continue
			}
			valv := reflect.ValueOf(val)
			if !valv.Type().ConvertibleTo(f.Type) {
				return fmt.Errorf("type mismatch in field %s. Cannot convert %s to %s", f.Name, valv.Type(), f.Type)
			}
			intov.Field(i).Set(valv.Convert(f.Type))
		}
	}
	return nil
}
