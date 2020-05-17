package simulator

import (
	"strings"
	"testing"
	"time"

	"github.com/edwardbrowncross/amazon-connect-simulator/event"
)

// TestHelper provides utility methods for asserting behavior of an ongoing call.
type TestHelper struct {
	t    *testing.T
	c    Call
	evts <-chan event.Event
}

// NewTestHelper creates a new TestHelper wrapping an ongoing call.
// TestHelper provides utility methods for asserting behavior of an ongoing call.
func NewTestHelper(t *testing.T, c Call) TestHelper {
	evts := make(chan event.Event)
	go func() {
		es := make([]event.Event, 0)
		outCh := func() chan<- event.Event {
			if len(es) == 0 {
				return nil
			}
			return evts
		}
		eNext := func() event.Event {
			if len(es) == 0 {
				return nil
			}
			return es[0]
		}
		for {
			select {
			case <-c.O:
			case e := <-c.Evt:
				es = append(es, e)
			case outCh() <- eNext():
				es = es[1:]
			}
		}
	}()
	return TestHelper{t, c, evts}
}

// ToEnter sends the given string as either a numeric entry or as an option selection.
// If not all characters can be sent, or more characters are required, it errors the test.
func (th *TestHelper) ToEnter(input string) {
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
		return
	}
}

// Message asserts that the next speech output matches the given string.
// If the simulator does not output anything after a second, it is assumed it is stuck and it fatals the test.
func (th *TestHelper) Message(msg string) {
	for {
		select {
		case evt := <-th.evts:
			switch evt.Type() {
			case event.PromptType:
				if got := evt.(event.PromptEvent).Text; got != msg {
					th.t.Errorf("expected to hear message '%s' but got '%s'", msg, got)
				}
				return
			}
		case <-time.After(time.Second):
			th.t.Fatalf("expected to hear message '%s' but it timed out", msg)
			return
		}
	}
}

// MessageContaining asserts that the next speech output contains the given string.
// If the simulator does not output anything after a second, it is assumed it is stuck and it fatals the test.
func (th *TestHelper) MessageContaining(msg string) {
	select {
	case evt := <-th.evts:
		switch evt.Type() {
		case event.PromptType:
			if got := evt.(event.PromptEvent).Text; !strings.Contains(got, msg) {
				th.t.Errorf("expected to hear message containing '%s' but got '%s'", msg, got)
			}
		}
	case <-time.After(time.Second):
		th.t.Fatalf("expected to hear message containing '%s' but it timed out", msg)
	}
}
