package simulator

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/edwardbrowncross/amazon-connect-simulator/event"
	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
	"github.com/edwardbrowncross/amazon-connect-simulator/module"
	"github.com/google/uuid"
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
	var contactID string
	if uuid, err := uuid.NewUUID(); err != nil {
		contactID = uuid.String()
	}
	c.System[flow.SystemCustomerNumber] = conf.DestNumber
	c.System[flow.SystemDialedNumber] = conf.SourceNumber
	c.System[flow.SystemChannel] = "VOICE"
	c.System[flow.SystemInitiationMethod] = "INBOUND"
	c.System[flow.SystemContactID] = contactID
	c.System[flow.SystemPreviousContactID] = contactID
	c.System[flow.SystemInitialContactID] = contactID
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
			m := cs.GetModule(*next)
			if m == nil {
				err = fmt.Errorf("missing module: %v", *next)
			}
			c.emit(event.NewModuleEvent(*m))
			next, err = module.MakeRunner(*m).Run(&cs)
		}
	}
	c.emit(event.DisconnectEvent{})
	c.Err = err
	close(c.o)
	c.evtsMutex.Lock()
	for _, ch := range c.evts {
		close(ch)
	}
	c.evtsMutex.Unlock()
}

func (c *Call) emit(event event.Event) {
	c.evtsMutex.Lock()
	for _, evt := range c.evts {
		evt <- event
	}
	c.evtsMutex.Unlock()
}

// Subscribe registers to receive structured events from the call.
// It takes a channel which events will be written to.
// The call will be blocked if the events cannot be written to the channel.
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

func (s *callConnector) Send(msg string, ssml bool) {
	s.emit(event.PromptEvent{
		Text: msg,
		SSML: ssml,
	})
	s.o <- msg
}

// Receive waits for a number of characters to be input.
// If the first character is not received before the timeout time, it returns nil.
func (s *callConnector) Receive(maxDigits int, timeout time.Duration, encrypt bool) *string {
	s.emit(event.InputEvent{
		MaxDigits: maxDigits,
		Timeout:   timeout,
	})
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
	for len(got) < maxDigits && got[len(got)-1] != '#' {
		got = append(got, <-s.i)
	}
	if got[len(got)-1] == '#' {
		got = got[:len(got)-1]
	}

	r := string(got)
	if encrypt {
		r = s.Encrypt(r)
	}
	return &r
}

// SetExternal sets a value into the state machine.
func (s *callConnector) SetExternal(key string, value interface{}) {
	s.External[key] = fmt.Sprintf("%v", value)
}

// SetContactData sets a value into the state machine.
func (s *callConnector) SetContactData(key string, value interface{}) {
	s.emit(event.UpdateContactDataEvent{
		Key:   key,
		Value: fmt.Sprintf("%v", value),
	})
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
	s.emit(event)
}

func (s *callConnector) InvokeLambda(named string, params json.RawMessage) (out string, outErr error, err error) {
	attr, _ := json.Marshal(s.ContactData)
	payloadIn := LambdaPayload{
		Details: lambdaPayloadDetails{
			ContactData: lambdaPayloadContactData{
				Attributes: attr,
				Channel:    s.System[flow.SystemChannel],
				CustomerEndpoint: lambdaPayloadContactEndpoint{
					Type:    "TELEPHONE_NUMBER",
					Address: s.System[flow.SystemCustomerNumber],
				},
				SystemEndpoint: lambdaPayloadContactEndpoint{
					Type:    "TELEPHONE_NUMBER",
					Address: s.System[flow.SystemDialedNumber],
				},
				InitiationMethod:  s.System[flow.SystemInitiationMethod],
				ContactID:         s.System[flow.SystemContactID],
				PreviousContactID: s.System[flow.SystemPreviousContactID],
				InitialContactID:  s.System[flow.SystemInitialContactID],
				Queue:             s.GetSystem(flow.SystemQueueARN),
			},
			Parameters: params,
		},
		Name: "ContactFlowEvent",
	}
	jsonIn, _ := json.Marshal(payloadIn)
	s.emit(event.InvokeLambdaEvent{
		ARN:         named,
		ParamsJSON:  string(params),
		PayloadJSON: string(jsonIn),
	})
	return s.simulatorConnector.InvokeLambda(named, string(jsonIn))
}
