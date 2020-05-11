package call

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

type Runner interface {
	Run(*Context) (*flow.ModuleID, error)
}

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

func (ctx *Context) Send(s string) {
	ctx.o <- s
}

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

type KeyValue struct {
	K string
	V string
}

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

func (ctx *Context) GetLambda(named string) interface{} {
	return ctx.getLambda(named)
}

func (ctx *Context) GetFlowStart(flowName string) *flow.ModuleID {
	return ctx.getFlowStart(flowName)
}
