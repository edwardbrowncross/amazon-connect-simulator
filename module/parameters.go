package module

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"

	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

type valueGetter interface {
	GetExternal(key string) interface{}
	GetContactData(key string) interface{}
	GetSystem(key string) interface{}
}

// parameterResolver uses the base methods of the CallConnector to perform more sophisticated lookup operations.
type parameterResolver struct {
	valueGetter
}

// get gets a single value by namespace and key.
func (call parameterResolver) get(namespace flow.ModuleParameterNamespace, key string) (interface{}, error) {
	switch namespace {
	case flow.NamespaceUserDefined:
		return call.GetContactData(key), nil
	case flow.NamespaceExternal:
		return call.GetExternal(key), nil
	case flow.NamespaceSystem:
		return call.GetSystem(key), nil
	default:
		return nil, fmt.Errorf("unknown namespace: %s", namespace)
	}
}

// resolve takes a raw module parameter and looks up its value (whether static or dynamic).
func (call parameterResolver) resolve(p flow.ModuleParameter) (val interface{}, err error) {
	if p.Namespace == nil || *p.Namespace == "" {
		val = p.Value
	} else {
		key, ok := p.Value.(string)
		if !ok {
			return
		}
		val, err = call.get(*p.Namespace, key)
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
func (call parameterResolver) unmarshal(plist flow.ModuleParameterList, into interface{}) error {
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
				val, err := call.resolve(p)
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
			val, err := call.resolve(*p)
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

var jsonP = regexp.MustCompile(`\$\.([a-zA-Z]+)\.([0-9a-zA-Z_\-]+)`)

// jsonPath takes a string like "you live in $.External.city" and interpolates the jsonPath components.
func (call parameterResolver) jsonPath(msg string) (out string) {
	out = jsonP.ReplaceAllStringFunc(msg, func(path string) (res string) {
		bits := jsonP.FindSubmatch([]byte(path))
		namespace := string(bits[1])
		key := string(bits[2])
		var val interface{}
		switch namespace {
		case "Attributes":
			val = call.GetContactData(key)
		case "External":
			val = call.GetExternal(key)
		default:
			val = path
		}
		return fmt.Sprintf("%v", val)
	})
	return
}