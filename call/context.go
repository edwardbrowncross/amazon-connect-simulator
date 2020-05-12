package call

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

// Runner takes a call context and returns the ID of the next block to run, or nil if the call is over.
type Runner interface {
	Run(*Context) (*flow.ModuleID, error)
}

// Context is the internal state machine behind a call.
type Context struct {
	o            chan<- string
	i            <-chan rune
	External     map[string]string
	ContactData  map[string]string
	System       map[flow.SystemKey]string
	getLambda    func(named string) interface{}
	getFlowStart func(named string) *flow.ModuleID
	getRunner    func(withID flow.ModuleID) Runner
}

// Send sends spoken text to the speaker.
func (ctx *Context) Send(s string) {
	ctx.o <- s
}

// Receive waits for a number of characters to be input.
// If the first character is not received before the timeout time, it returns nil.
func (ctx *Context) Receive(count int, timeout time.Duration) *string {
	got := []rune{}
	select {
	case <-time.After(timeout):
		return nil
	case in := <-ctx.i:
		got = append(got, in)
	}
	for len(got) < count {
		got = append(got, <-ctx.i)
	}
	r := string(got)
	return &r
}

// ResolveValue looks up a dynamic value in the context's state machine with namspaces and key.
func (ctx *Context) ResolveValue(namespace flow.ModuleParameterNamespace, key string) interface{} {
	switch namespace {
	case flow.NamespaceUserDefined:
		return ctx.ContactData[key]
	case flow.NamespaceExternal:
		return ctx.External[key]
	case flow.NamespaceSystem:
		return ctx.System[flow.SystemKey(key)]
	}
	return nil
}

// KeyValue represents the parsed value of key-value parameter.
type KeyValue struct {
	K string
	V string
}

// ResolveParameter takes a raw module parameter and looks up its value (whether static or dynamic).
func (ctx *Context) ResolveParameter(p flow.ModuleParameter) interface{} {
	var val interface{}
	if p.Namespace == nil || *p.Namespace == "" {
		val = p.Value
	} else {
		key, ok := p.Value.(string)
		if !ok {
			return nil
		}
		val = ctx.ResolveValue(*p.Namespace, key)
	}
	if p.Key != "" {
		return KeyValue{
			K: p.Key,
			V: fmt.Sprintf("%v", val),
		}
	}
	return val
}

// UnmarshalParameters takes the list of a block's parameters and unmarshals it into a typed struct.
// The struct field names should match the names of the parameters.
// Type checking will be performed. If the type of the value cannot be converted to the field type, it errors.
// Where there are multiple parameters with the same name, use a field with a slice type.
func (ctx *Context) UnmarshalParameters(plist flow.ModuleParameterList, into interface{}) error {
	if reflect.ValueOf(into).Kind() != reflect.Ptr || into == nil {
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
				val := ctx.ResolveParameter(p)
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
			val := ctx.ResolveParameter(*p)
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

// GetLambda looks up a lambda in the base connect configuration.
func (ctx *Context) GetLambda(named string) interface{} {
	return ctx.getLambda(named)
}

// GetFlowStart looks up the ID of the block that starts the flow with the given name.
func (ctx *Context) GetFlowStart(flowName string) *flow.ModuleID {
	return ctx.getFlowStart(flowName)
}
