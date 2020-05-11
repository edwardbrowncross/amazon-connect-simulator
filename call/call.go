package call

import "github.com/edwardbrowncross/amazon-connect-simulator/flow"

type Call struct {
	O   <-chan string
	I   chan<- rune
	ctx Context
}

type FlowDescriber struct {
	GetLambda    func(named string) interface{}
	GetFlowStart func(withName string) *flow.ModuleID
	GetRunner    func(withID flow.ModuleID) Runner
}

type Config struct {
	SourceNumber string
	DestNumber   string
}

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
