package simulator

import (
	"strings"
	"testing"
	"time"
)

// TestHelper provides utility methods for asserting behavior of an ongoing call.
type TestHelper struct {
	t *testing.T
	c Call
}

// NewTestHelper creates a new TestHelper wrapping an ongoing call.
// TestHelper provides utility methods for asserting behavior of an ongoing call.
func NewTestHelper(t *testing.T, c Call) TestHelper {
	return TestHelper{t, c}
}

// ToEnter sends the given string as either a numeric entry or as an option selection.
// If not all characters can be sent, it errors the test.
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
	select {
	case got := <-th.c.O:
		if got != msg {
			th.t.Errorf("expected to hear message '%s' but got '%s'", msg, got)
		}
	case <-time.After(time.Second):
		th.t.Fatalf("expected to hear message '%s' but it timed out", msg)
	}
}

// MessageContaining asserts that the next speech output contains the given string.
// If the simulator does not output anything after a second, it is assumed it is stuck and it fatals the test.
func (th *TestHelper) MessageContaining(msg string) {
	select {
	case got := <-th.c.O:
		if !strings.Contains(got, msg) {
			th.t.Errorf("expected to hear message '%s' but got '%s'", msg, got)
		}
	case <-time.After(time.Second):
		th.t.Fatalf("expected to hear message '%s' but it timed out", msg)
	}
}
