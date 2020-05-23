package simulator

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/edwardbrowncross/amazon-connect-simulator/event"
)

// TestHelper provides utility methods for asserting behavior of an ongoing call.
type TestHelper struct {
	t           *testing.T
	c           *Call
	evts        []event.Event
	ready       <-chan bool
	readyToggle chan<- bool
	mutex       sync.RWMutex
}

// NewTestHelper creates a new TestHelper wrapping an ongoing call.
// TestHelper provides utility methods for asserting behavior of an ongoing call.
func NewTestHelper(t *testing.T, c *Call) *TestHelper {
	buffer := make(chan event.Event, 64)
	c.Subscribe(buffer)
	readyVal, readyToggle := toggleChannel()

	th := TestHelper{
		t:           t,
		c:           c,
		ready:       readyVal,
		readyToggle: readyToggle,
		evts:        make([]event.Event, 0),
	}

	go func() {
		for {
			select {
			case <-c.O:
			case evt, ok := <-buffer:
				if !ok {
					close(readyToggle)
					return
				}
				th.mutex.Lock()
				th.evts = append(th.evts, evt)
				th.mutex.Unlock()
				switch evt.Type() {
				case event.DisconnectType, event.InputType, event.TransferQueueType:
					readyToggle <- true
				case event.ModuleType:
					fallthrough
				default:
					readyToggle <- false
				}
			}
		}
	}()

	return &th
}

func (th *TestHelper) readEvents() (ok bool) {
	_, ok = <-th.ready
	return
}

func (th *TestHelper) cancelReady() {
	th.readyToggle <- false
}

func (th *TestHelper) run(m matcher) {
	ok := th.readEvents()
	th.mutex.RLock()
	defer th.mutex.RUnlock()
	var gots []string
	for i, e := range th.evts {
		match, pass, got := m.match(e)
		if !match {
			continue
		}
		if pass {
			th.evts = th.evts[i:]
			return
		}
		gots = append(gots, got)
	}
	if len(gots) == 0 {
		gots = append(gots, "nothing")
	}
	th.t.Errorf("expected %s. Got: \n%v.", m.expected(), strings.Join(gots, "\n"))
	if !ok {
		if th.c.Err != nil {
			th.t.Fatalf("call ended with error: %v", th.c.Err)
		} else {
			th.t.Fatal("call terminated unexpectedly")
}
	}
}

// ToEnter sends the given string as either a numeric entry or as an option selection.
// If not all characters can be sent, or more characters are required, it errors the test.
func (th *TestHelper) ToEnter(input string) {
	th.cancelReady()
	for i, r := range input {
		select {
		case th.c.I <- r:
			continue
		case <-time.After(time.Second):
			th.t.Errorf("expected to be able to send input %s, but was only able to send %d characters", input, i)
		}
	}
	select {
	case th.c.I <- '#':
		th.t.Errorf("expected input of %s to fill the input, but it did not.", input)
	default:
		break
	}
}

// Message asserts that one of the following speech output matches the given string.
func (th *TestHelper) Message(msg string) {
	th.run(promptExactMatcher{msg})
}

// MessageContaining asserts that one of the following speech output contains the given string.
func (th *TestHelper) MessageContaining(msg string) {
	th.run(promptPartialMatcher{msg})
}

// TransferToQueue asserts that the caller is transferred to a queue with the given name.
func (th *TestHelper) TransferToQueue(named string) {
	th.run(queueTransferMatcher{named})
}

// TransferToFlow asserts that the call moved to the flow with the given name.
func (th *TestHelper) TransferToFlow(named string) {
	th.run(flowTransferMatcher{named})
}

type matcher interface {
	match(event.Event) (match bool, pass bool, got string)
	expected() string
}

type promptExactMatcher struct {
	text string
}

func (m promptExactMatcher) match(evt event.Event) (match bool, pass bool, got string) {
	if evt.Type() != event.PromptType {
		return false, false, ""
	}
	e := evt.(event.PromptEvent)
	match = true
	got = e.Text
	pass = bool(e.Text == m.text)
	return
}

func (m promptExactMatcher) expected() string {
	return fmt.Sprintf("to get prompt of exactly '%s'", m.text)
}

type promptPartialMatcher struct {
	text string
}

func (m promptPartialMatcher) match(evt event.Event) (match bool, pass bool, got string) {
	if evt.Type() != event.PromptType {
		return false, false, ""
	}
	e := evt.(event.PromptEvent)
	match = true
	got = e.Text
	pass = bool(strings.Contains(e.Text, m.text))
	return
}

func (m promptPartialMatcher) expected() string {
	return fmt.Sprintf("to get prompt containing '%s'", m.text)
}

type queueTransferMatcher struct {
	queueName string
}

func (m queueTransferMatcher) match(evt event.Event) (match bool, pass bool, got string) {
	if evt.Type() != event.TransferQueueType {
		return false, false, ""
	}
	e := evt.(event.QueueTransferEvent)
	match = true
	got = e.QueueName
	pass = bool(e.QueueName == m.queueName)
	return
}

func (m queueTransferMatcher) expected() string {
	return fmt.Sprintf("to be transfered to queue '%s'", m.queueName)
}

type flowTransferMatcher struct {
	flowName string
}

func (m flowTransferMatcher) match(evt event.Event) (match bool, pass bool, got string) {
	if evt.Type() != event.TransferFlowType {
		return false, false, ""
	}
	e := evt.(event.FlowTransferEvent)
	match = true
	got = e.FlowName
	pass = bool(e.FlowName == m.flowName)
	return
}

func (m flowTransferMatcher) expected() string {
	return fmt.Sprintf("to be transfered to flow '%s'", m.flowName)
}

func toggleChannel() (value <-chan bool, toggle chan<- bool) {
	v := make(chan bool)
	t := make(chan bool)
	go func() {
		ready := false
		for {
			if !ready {
			falseLoop:
				for {
					r, ok := <-t
					if !ok {
						close(v)
						return
					}
					if r {
						ready = true
						break falseLoop
					}
				}
			} else {
			trueLoop:
				for {
					select {
					case v <- true:
					case r, ok := <-t:
						if !ok {
							close(v)
							return
						}
						if !r {
							ready = false
							break trueLoop
						}
					}
				}
			}
		}
	}()
	return v, t
}
