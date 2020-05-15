package call

import (
	"fmt"
	"time"

	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
	"github.com/edwardbrowncross/amazon-connect-simulator/module"
)

// Context is the internal state machine behind a call.
type Context struct {
	o            chan<- string
	i            <-chan rune
	External     map[string]string
	ContactData  map[string]string
	System       map[string]string
	getLambda    func(named string) interface{}
	getFlowStart func(named string) *flow.ModuleID
	getRunner    func(withID flow.ModuleID) module.Runner
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

// SetExternal sets a value into the context's state machine.
func (ctx *Context) SetExternal(key string, value interface{}) {
	ctx.External[key] = fmt.Sprintf("%v", value)
}

// SetContactData sets a value into the context's state machine.
func (ctx *Context) SetContactData(key string, value interface{}) {
	ctx.ContactData[key] = fmt.Sprintf("%v", value)
}

// SetSystem sets a value into the context's state machine.
func (ctx *Context) SetSystem(key string, value interface{}) {
	ctx.System[key] = fmt.Sprintf("%v", value)
}

// GetExternal gets a value from the context's state machine.
func (ctx *Context) GetExternal(key string) interface{} {
	val, found := ctx.External[key]
	if !found {
		return nil
	}
	return val
}

// GetContactData gets a value from the context's state machine.
func (ctx *Context) GetContactData(key string) interface{} {
	val, found := ctx.ContactData[key]
	if !found {
		return nil
	}
	return val
}

// GetSystem gets a value from the context's state machine.
func (ctx *Context) GetSystem(key string) interface{} {
	val, found := ctx.System[key]
	if !found {
		return nil
	}
	return val
}

// ClearExternal allows clearing of all externalvalues in the state machine.
func (ctx *Context) ClearExternal() {
	ctx.External = map[string]string{}
}

// GetLambda looks up a lambda in the base connect configuration.
func (ctx *Context) GetLambda(named string) interface{} {
	return ctx.getLambda(named)
}

// GetFlowStart looks up the ID of the block that starts the flow with the given name.
func (ctx *Context) GetFlowStart(flowName string) *flow.ModuleID {
	return ctx.getFlowStart(flowName)
}
