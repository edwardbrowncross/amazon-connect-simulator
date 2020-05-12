package call

import "github.com/edwardbrowncross/amazon-connect-simulator/flow"

// Call is used to interact with an ongoing call.
type Call struct {
	// Output (speaker).
	O <-chan string
	// Input (keypad).
	I   chan<- rune
	ctx Context
}

// FlowDescriber allows a call to get information about the flow controlling it.
type FlowDescriber struct {
	GetLambda    func(named string) interface{}
	GetFlowStart func(withName string) *flow.ModuleID
	GetRunner    func(withID flow.ModuleID) Runner
}

// Config is data unique to this particular call.
type Config struct {
	SourceNumber string
	DestNumber   string
}

// New is used by the simulator to create a new call.
func New(conf Config, fd FlowDescriber, start flow.ModuleID) Call {
	out := make(chan string)
	in := make(chan rune)
	ctx := Context{
		o:            out,
		i:            in,
		External:     map[string]string{},
		ContactData:  map[string]string{},
		System:       map[flow.SystemKey]string{},
		getLambda:    fd.GetLambda,
		getFlowStart: fd.GetFlowStart,
		getRunner:    fd.GetRunner,
	}
	c := Call{
		O:   out,
		I:   in,
		ctx: ctx,
	}
	go c.run(start)
	return c
}

func (c *Call) run(start flow.ModuleID) {
	var next *flow.ModuleID
	var err error
	next = &start
	for next != nil && err == nil {
		m := c.ctx.getRunner(*next)
		next, err = m.Run(&c.ctx)
	}
}
