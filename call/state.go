package call

import (
	"fmt"
	"time"
)

// State is the internal state machine behind a call.
type State struct {
	SimulatorConnector
	o           chan<- string
	i           <-chan rune
	External    map[string]string
	ContactData map[string]string
	System      map[string]string
}

// Send sends spoken text to the speaker.
func (s *State) Send(msg string) {
	s.o <- msg
}

// Receive waits for a number of characters to be input.
// If the first character is not received before the timeout time, it returns nil.
func (s *State) Receive(count int, timeout time.Duration) *string {
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
func (s *State) SetExternal(key string, value interface{}) {
	s.External[key] = fmt.Sprintf("%v", value)
}

// SetContactData sets a value into the state machine.
func (s *State) SetContactData(key string, value interface{}) {
	s.ContactData[key] = fmt.Sprintf("%v", value)
}

// SetSystem sets a value into the state machine.
func (s *State) SetSystem(key string, value interface{}) {
	s.System[key] = fmt.Sprintf("%v", value)
}

// GetExternal gets a value from the state machine.
func (s *State) GetExternal(key string) interface{} {
	val, found := s.External[key]
	if !found {
		return nil
	}
	return val
}

// GetContactData gets a value from the state machine.
func (s *State) GetContactData(key string) interface{} {
	val, found := s.ContactData[key]
	if !found {
		return nil
	}
	return val
}

// GetSystem gets a value from the state machine.
func (s *State) GetSystem(key string) interface{} {
	val, found := s.System[key]
	if !found {
		return nil
	}
	return val
}

// ClearExternal allows clearing of all externalvalues in the state machine.
func (s *State) ClearExternal() {
	s.External = map[string]string{}
}
