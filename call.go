package simulator

import (
	"fmt"
	"sync"
	"time"

	"github.com/edwardbrowncross/amazon-connect-simulator/event"
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
	Err         error
	evts        []chan<- event.Event
	kill        chan<- interface{}
	evtsMutex   sync.Mutex
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
func newCall(conf CallConfig, sc *simulatorConnector, start flow.ModuleID) *Call {
	out := make(chan string)
	in := make(chan rune)
	kill := make(chan interface{})
	c := Call{
		O:           out,
		I:           in,
		o:           out,
		i:           in,
		kill:        kill,
		evtsMutex:   sync.Mutex{},
		evts:        make([]chan<- event.Event, 0),
		External:    map[string]string{},
		ContactData: map[string]string{},
		System:      map[string]string{},
	}
	go c.run(start, callConnector{&c, sc}, kill)
	return &c
}

func (c *Call) run(start flow.ModuleID, cs callConnector, kill <-chan interface{}) {
	var next *flow.ModuleID
	var err error
	next = &start
loop:
	for next != nil && err == nil {
		select {
		case _, ok := <-kill:
			if !ok {
				break loop
			}
		default:
			m := cs.GetRunner(*next)
			next, err = m.Run(&cs)
		}
	}
	c.Err = err
	close(c.o)
	c.evtsMutex.Lock()
	for _, ch := range c.evts {
		close(ch)
	}
	c.evtsMutex.Unlock()
}

// Subscribe registers to receive structured events from the call.
// It takes a channel which events will be written to (without blocking the call).
func (c *Call) Subscribe(events chan<- event.Event) {
	c.evtsMutex.Lock()
	c.evts = append(c.evts, events)
	c.evtsMutex.Unlock()
}

// Terminate ends an ongoing call.
// If the call has already ended, it does nothing.
func (c *Call) Terminate() {
	close(c.kill)
}

// callConnector exposes methods for modules to interact with the ongoing call.
type callConnector struct {
	*Call
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
	case in, ok := <-s.i:
		if !ok {
			s.Terminate()
		}
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

func (s *callConnector) Emit(event event.Event) {
	s.evtsMutex.Lock()
	for _, evt := range s.evts {
		select {
		case evt <- event:
		default:
		}
	}
	s.evtsMutex.Unlock()
}
