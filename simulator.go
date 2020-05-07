package simulator

import (
	"amazon-connect-simulator/connect"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type SystemKey string

const (
	SystemLastUserInput SystemKey = "Stored customer input"
	SystemQueueName               = "Queue name"
	SystemQueueARN                = "Queue ARN"
)

type lambdaHandler func(context.Context, interface{}) (interface{}, error)

type runnable interface {
	run(*CallSimulator) (*ModuleID, error)
}

type CallSimulator struct {
	O           <-chan string
	I           chan<- rune
	o           chan<- string
	i           <-chan rune
	external    map[string]string
	contactData map[string]string
	system      map[SystemKey]string
	lambdas     map[string]interface{}
	flows       map[string]Flow
	modules     map[ModuleID]runnable
}

func NewCallSimulator() CallSimulator {
	out := make(chan string)
	in := make(chan rune)
	return CallSimulator{
		O:           out,
		I:           in,
		o:           out,
		i:           in,
		external:    map[string]string{},
		contactData: map[string]string{},
		system:      map[SystemKey]string{},
		lambdas:     map[string]interface{}{},
		flows:       map[string]Flow{},
		modules:     map[ModuleID]runnable{},
	}
}

func (cs *CallSimulator) LoadFlow(flow Flow) {
	cs.flows[flow.Metadata.Name] = flow
	for _, m := range flow.Modules {
		var r runnable
		switch m.Type {
		case ModuleStoreUserInput:
			r = storeUserInput(m)
		case ModuleCheckAttribute:
			r = checkAttribute(m)
		case ModuleTransfer:
			r = transfer(m)
		case ModulePlayPrompt:
			r = playPrompt(m)
		case ModuleDisconnect:
			r = disconnect(m)
		case ModuleSetQueue:
			r = setQueue(m)
		case ModuleGetUserInput:
			r = getUserInput(m)
		case ModuleSetAttributes:
			r = setAttributes(m)
		case ModuleInvokeExternalResource:
			r = invokeExternalResource(m)
		case ModuleCheckHoursOfOperation:
			r = checkHoursOfOperation(m)
		default:
			r = passthrough(m)
		}
		cs.modules[m.ID] = r
	}
}

func (cs *CallSimulator) StartCall(entryFlowName string) (err error) {
	f, ok := cs.flows[entryFlowName]
	if !ok {
		return errors.New("entry flow not found")
	}
	var mid *ModuleID
	mid = &f.Start
	for err == nil && mid != nil {
		// fmt.Println(*mid)
		mid, err = cs.modules[*mid].run(cs)
	}
	return
}

func (cs *CallSimulator) RegisterLambda(name string, fn interface{}) error {
	err := validateLambda(fn)
	if err != nil {
		return err
	}
	cs.lambdas[name] = fn
	return nil
}

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
		return fmt.Errorf("expected second argument to be struct but it was %s", fnt.In(1).Kind())
	}
	if fnt.NumOut() != 2 {
		return fmt.Errorf("expected function to return 2 elements but it returns %d", fnt.NumOut())
	}
	if fnt.Out(0).Kind() != reflect.Struct {
		return fmt.Errorf("expected first return to be struct but it was %s", fnt.Out(0).Kind())
	}
	errort := reflect.TypeOf((*error)(nil)).Elem()
	if !fnt.Out(1).Implements(errort) {
		return errors.New("expected seclnd return to be an error")
	}
	return nil
}

func (cs *CallSimulator) send(s string) {
	cs.o <- s
}

func (cs *CallSimulator) receive(count int, timeout time.Duration) *string {
	got := []rune{}
	select {
	case <-time.After(timeout):
		return nil
	case in := <-cs.i:
		got = append(got, in)
	}
	for len(got) < count {
		got = append(got, <-cs.i)
	}
	r := string(got)
	return &r
}

func (cs *CallSimulator) resolveValue(namespace ModuleParameterNamespace, key string) interface{} {
	switch namespace {
	case NamespaceUserDefined:
		return cs.contactData[key]
	case NamespaceExternal:
		return cs.external[key]
	case NamespaceSystem:
		return cs.system[SystemKey(key)]
	}
	return nil
}

type keyValue struct {
	k string
	v string
}

func (cs *CallSimulator) resolveParameter(p ModuleParameter) interface{} {
	var val interface{}
	if p.Namespace == nil || *p.Namespace == "" {
		val = p.Value
	} else {
		key, ok := p.Value.(string)
		if !ok {
			return nil
		}
		val = cs.resolveValue(*p.Namespace, key)
	}
	if p.Key != "" {
		return keyValue{
			k: p.Key,
			v: fmt.Sprintf("%v", val),
		}
	}
	return val
}

func (cs *CallSimulator) unmarshalParameters(plist ModuleParameterList, into interface{}) error {
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
				val := cs.resolveParameter(p)
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
			val := cs.resolveParameter(*p)
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

func (cs *CallSimulator) lookupLambda(arn string) *interface{} {
	for k, v := range cs.lambdas {
		if strings.Contains(arn, k) {
			return &v
		}
	}
	return nil
}

type storeUserInput Module

type storeUserInputParams struct {
	Text      string
	Timeout   string
	MaxDigits int
}

func (m storeUserInput) run(cs *CallSimulator) (next *ModuleID, err error) {
	if m.Type != ModuleStoreUserInput {
		return nil, fmt.Errorf("module of type %s being run as storeUserInput", m.Type)
	}
	p := storeUserInputParams{}
	err = cs.unmarshalParameters(m.Parameters, &p)
	if err != nil {
		return
	}
	cs.send(p.Text)
	timeout, err := strconv.Atoi(p.Timeout)
	if err != nil {
		return
	}
	entry := cs.receive(p.MaxDigits, time.Duration(timeout)*time.Second)
	if entry == nil {
		next = m.Branches.GetLink(BranchError)
		return
	}
	cs.system[SystemLastUserInput] = *entry
	next = m.Branches.GetLink(BranchSuccess)
	return
}

func evaluateConditions(c ModuleBranchList, v string) (*ModuleID, error) {
	conditions := c.List(BranchEvaluate)
	for _, c := range conditions {
		pass := false
		switch c.ConditionType {
		case ConditionEquals:
			pass = bool(v == c.ConditionValue)
		case ConditionGT:
			pass = bool(v > c.ConditionValue)
		case ConditionGTE:
			pass = bool(v >= c.ConditionValue)
		case ConditionLT:
			pass = bool(v < c.ConditionValue)
		case ConditionLTE:
			pass = bool(v <= c.ConditionValue)
		default:
			return nil, fmt.Errorf("unhandled condition type: %s", c.ConditionType)
		}
		if pass {
			return &c.Transition, nil
		}
	}
	return c.GetLink(BranchNoMatch), nil
}

type checkAttribute Module

type checkAttributeParams struct {
	Namespace string
	Attribute string
}

func (m checkAttribute) run(cs *CallSimulator) (next *ModuleID, err error) {
	if m.Type != ModuleCheckAttribute {
		return nil, fmt.Errorf("module of type %s being run as checkAttribute", m.Type)
	}
	p := checkAttributeParams{}
	err = cs.unmarshalParameters(m.Parameters, &p)
	if err != nil {
		return
	}
	v := cs.resolveValue(ModuleParameterNamespace(p.Namespace), p.Attribute)
	vs := fmt.Sprintf("%s", v)
	return evaluateConditions(m.Branches, vs)
}

type transfer Module

func (m transfer) run(cs *CallSimulator) (next *ModuleID, err error) {
	if m.Type != ModuleTransfer {
		return nil, fmt.Errorf("module of type %s being run as transfer", m.Type)
	}
	switch m.Target {
	case TargetFlow:
		fName := m.Parameters.Get("ContactFlowId").ResourceName
		f := cs.flows[fName]
		return &f.Start, nil
	case TargetQueue:
		cs.send(fmt.Sprintf("<transfer to queue %s>", cs.system[SystemQueueName]))
		return nil, nil
	default:
		return nil, fmt.Errorf("unhandled transfer target: %s", m.Target)
	}
}

type playPrompt Module

type playPromptParams struct {
	Text string
}

func (m playPrompt) run(cs *CallSimulator) (next *ModuleID, err error) {
	if m.Type != ModulePlayPrompt {
		return nil, fmt.Errorf("module of type %s being run as playPrompt", m.Type)
	}
	p := playPromptParams{}
	err = cs.unmarshalParameters(m.Parameters, &p)
	if err != nil {
		return
	}
	cs.send(p.Text)
	return m.Branches.GetLink(BranchSuccess), nil
}

type disconnect Module

func (m disconnect) run(cs *CallSimulator) (next *ModuleID, err error) {
	if m.Type != ModuleDisconnect {
		return nil, fmt.Errorf("module of type %s being run as disconnect", m.Type)
	}
	return nil, nil
}

type setQueue Module

func (m setQueue) run(cs *CallSimulator) (next *ModuleID, err error) {
	if m.Type != ModuleSetQueue {
		return nil, fmt.Errorf("module of type %s being run as setQueue", m.Type)
	}
	p := m.Parameters.Get("Queue")
	if p == nil {
		return nil, errors.New("missing Queue parameter")
	}
	cs.system[SystemQueueARN] = p.Value.(string)
	cs.system[SystemQueueName] = p.ResourceName
	return m.Branches.GetLink(BranchSuccess), nil
}

type getUserInput Module

type getUserInputParams struct {
	Text      string
	Timeout   string
	MaxDigits string
}

func (m getUserInput) run(cs *CallSimulator) (next *ModuleID, err error) {
	if m.Type != ModuleGetUserInput {
		return nil, fmt.Errorf("module of type %s being run as getUserInput", m.Type)
	}
	p := getUserInputParams{}
	err = cs.unmarshalParameters(m.Parameters, &p)
	if err != nil {
		return
	}
	cs.send(p.Text)
	md, err := strconv.Atoi(p.MaxDigits)
	if err != nil {
		return nil, fmt.Errorf("invalid MaxDigits: %s", p.MaxDigits)
	}
	tm, err := strconv.Atoi(p.Timeout)
	if err != nil {
		return nil, fmt.Errorf("invalid Timeout: %s", p.Timeout)
	}
	in := cs.receive(md, time.Duration(tm)*time.Second)
	if in == nil {
		return m.Branches.GetLink(BranchTimeout), nil
	}
	return evaluateConditions(m.Branches, *in)
}

type setAttributes Module

type setAttributesParams struct {
	Attribute []keyValue
}

func (m setAttributes) run(cs *CallSimulator) (next *ModuleID, err error) {
	if m.Type != ModuleSetAttributes {
		return nil, fmt.Errorf("module of type %s being run as setAttributes", m.Type)
	}
	p := setAttributesParams{}
	err = cs.unmarshalParameters(m.Parameters, &p)
	if err != nil {
		return
	}
	for _, a := range p.Attribute {
		cs.contactData[a.k] = a.v
	}
	return m.Branches.GetLink(BranchSuccess), nil
}

type invokeExternalResource Module

type invokeExternalResourceParams struct {
	TimeLimit   string
	FunctionArn string
	Parameter   []keyValue
}

func (m invokeExternalResource) run(cs *CallSimulator) (next *ModuleID, err error) {
	if m.Type != ModuleInvokeExternalResource {
		return nil, fmt.Errorf("module of type %s being run as invokeExternalResource", m.Type)
	}
	p := invokeExternalResourceParams{}
	err = cs.unmarshalParameters(m.Parameters, &p)
	if err != nil {
		return
	}
	fn := cs.lookupLambda(p.FunctionArn)
	if fn == nil {
		return nil, fmt.Errorf("no function found for lambda %s", p.FunctionArn)
	}
	fields := make([]string, len(p.Parameter))
	for i, p := range p.Parameter {
		v, _ := json.Marshal(p.v)
		fields[i] = fmt.Sprintf(`"%s":%s`, p.k, v)
	}
	paramsIn := fmt.Sprintf(`{%s}`, strings.Join(fields, ","))
	payloadIn := connect.Payload{
		Details: connect.PayloadDetails{
			Parameters: json.RawMessage(paramsIn),
		},
	}
	jsonIn, _ := json.Marshal(payloadIn)
	jsonOut, err := m.invoke(*fn, string(jsonIn))
	if err != nil {
		fmt.Println(paramsIn)
		fmt.Println("debug:" + err.Error())
		return m.Branches.GetLink(BranchError), nil
	}
	out := map[string]interface{}{}
	err = json.Unmarshal([]byte(jsonOut), &out)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal json from lambda: %s", jsonOut)
	}
	cs.external = map[string]string{}
	for k, v := range out {
		cs.external[k] = fmt.Sprintf("%v", v)
	}
	return m.Branches.GetLink(BranchSuccess), nil
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

type checkHoursOfOperation Module

func (m checkHoursOfOperation) run(cs *CallSimulator) (next *ModuleID, err error) {
	if m.Type != ModuleCheckHoursOfOperation {
		return nil, fmt.Errorf("module of type %s being run as checkHoursOfOperation", m.Type)
	}
	return m.Branches.GetLink(BranchTrue), nil
}

type passthrough Module

func (m passthrough) run(cs *CallSimulator) (next *ModuleID, err error) {
	return m.Branches.GetLink(BranchSuccess), nil
}
