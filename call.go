package simulator

import (
	"fmt"
	"time"

	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

// Call is used to interact with an ongoing call.
type Call struct {
	// Output (speaker).
	O <-chan string
	// Input (keypad).
	I           chan<- rune
	o           chan<- string
	i           <-chan rune
	External    map[string]string
	ContactData map[string]string
	System      map[string]string
}

// CallConfig is data unique to this particular call.
type CallConfig struct {
	SourceNumber string
	DestNumber   string
}

// New is used by the simulator to create a new call.
func newCall(conf CallConfig, sc *simulatorConnector, start flow.ModuleID) Call {
	out := make(chan string)
	in := make(chan rune)
	c := Call{
		O:           out,
		I:           in,
		o:           out,
		i:           in,
		External:    map[string]string{},
		ContactData: map[string]string{},
		System:      map[string]string{},
	}
	go c.run(start, callConnector{c, sc})
	return c
}

func (c *Call) run(start flow.ModuleID, cs callConnector) {
	var next *flow.ModuleID
	var err error
	next = &start
	for next != nil && err == nil {
		m := cs.GetRunner(*next)
		next, err = m.Run(&cs)
	}
}

// callConnector exposes methods for modules to interact with the ongoing call.
type callConnector struct {
	Call
	*simulatorConnector
}

func (s *callConnector) Send(msg string) {
	s.o <- msg
}

// Receive waits for a number of characters to be input.
// If the first character is not received before the timeout time, it returns nil.
func (s *callConnector) Receive(count int, timeout time.Duration) *string {
	got := []rune{}
	select {
	case <-time.After(timeout):
		return nil
	case in := <-s.i:
		got = append(got, in)
	}
	for len(got) < count {
		got = append(got, <-s.i)
	}
	r := string(got)
	return &r
}

// SetExternal sets a value into the state machine.
func (s *callConnector) SetExternal(key string, value interface{}) {
	s.External[key] = fmt.Sprintf("%v", value)
}

// SetContactData sets a value into the state machine.
func (s *callConnector) SetContactData(key string, value interface{}) {
	s.ContactData[key] = fmt.Sprintf("%v", value)
}

// SetSystem sets a value into the state machine.
func (s *callConnector) SetSystem(key string, value interface{}) {
	s.System[key] = fmt.Sprintf("%v", value)
}

// GetExternal gets a value from the state machine.
func (s *callConnector) GetExternal(key string) interface{} {
	val, found := s.External[key]
	if !found {
		return nil
	}
	return val
}

// GetContactData gets a value from the state machine.
func (s *callConnector) GetContactData(key string) interface{} {
	val, found := s.ContactData[key]
	if !found {
		return nil
	}
	return val
}

// GetSystem gets a value from the state machine.
func (s *callConnector) GetSystem(key string) interface{} {
	val, found := s.System[key]
	if !found {
		return nil
	}
	return val
}

// ClearExternal allows clearing of all externalvalues in the state machine.
func (s *callConnector) ClearExternal() {
	s.External = map[string]string{}
}
