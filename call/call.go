package call

import (
	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
	"github.com/edwardbrowncross/amazon-connect-simulator/module"
)

// Call is used to interact with an ongoing call.
type Call struct {
	// Output (speaker).
	O <-chan string
	// Input (keypad).
	I     chan<- rune
	state State
}

// SimulatorConnector allows a call to get information about the conneect instance it is part off.
type SimulatorConnector interface {
	GetLambda(named string) interface{}
	GetFlowStart(withName string) *flow.ModuleID
	GetRunner(withID flow.ModuleID) module.Runner
}

// Config is data unique to this particular call.
type Config struct {
	SourceNumber string
	DestNumber   string
}

// New is used by the simulator to create a new call.
func New(conf Config, fd SimulatorConnector, start flow.ModuleID) Call {
	out := make(chan string)
	in := make(chan rune)
	state := State{
		SimulatorConnector: fd,
		o:                  out,
		i:                  in,
		External:           map[string]string{},
		ContactData:        map[string]string{},
		System:             map[string]string{},
	}
	c := Call{
		O:     out,
		I:     in,
		state: state,
	}
	go c.run(start)
	return c
}

func (c *Call) run(start flow.ModuleID) {
	var next *flow.ModuleID
	var err error
	next = &start
	for next != nil && err == nil {
		m := c.state.GetRunner(*next)
		next, err = m.Run(&c.state)
	}
}
