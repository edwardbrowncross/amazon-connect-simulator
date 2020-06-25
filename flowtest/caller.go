package flowtest

import "time"

// CallerContext is returned from Expect.Caller()
type CallerContext struct {
	testContext
}

// ToEnter sends the given string as either a numeric entry or as an option selection.
// If not all characters can be sent, or more characters are required, it errors the test.
func (tc CallerContext) ToEnter(input string) {
	tc.t.Helper()
	tc.expect.cancelReady()
	for i, r := range input {
		select {
		case tc.expect.c.Caller.I <- r:
			continue
		case <-time.After(time.Second):
			tc.t.Errorf("expected to be able to send input %s, but was only able to send %d characters", input, i)
			return
		}
	}
	select {
	case <-time.After(time.Second):
		tc.t.Errorf("expected input of %s to fill the input, but it did not.", input)
		tc.expect.readyToggle <- true
	case <-tc.expect.ready:
		break
	}
}

// ToPress sends the given rune as an option selection at a menu.
// If the press cannot be sent, or more characters are required, it errors the test.
func (tc CallerContext) ToPress(input rune) {
	tc.t.Helper()
	tc.expect.cancelReady()
	select {
	case tc.expect.c.Caller.I <- input:
	case <-time.After(time.Second):
		tc.t.Errorf("expected to be able to send input %s, but the flow was not ready for input", string(input))
		return
	}
	select {
	case <-time.After(time.Second):
		tc.t.Errorf("expected input of %s to fill the input, but it did not.", string(input))
		tc.expect.readyToggle <- true
	case <-tc.expect.ready:
		break
	}
}

// ToWaitForTimeout waits for the current input block to time out.
func (tc CallerContext) ToWaitForTimeout() {
	tc.t.Helper()
	tc.expect.cancelReady()
	select {
	case <-time.After(time.Second):
		tc.t.Error("expected to wait for timeout, but no input was required")
		return
	case tc.expect.c.Caller.I <- 'T':
	}
}
