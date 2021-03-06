package flowtest

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	simulator "github.com/edwardbrowncross/amazon-connect-simulator"
	"github.com/edwardbrowncross/amazon-connect-simulator/event"
)

// Expect provides utility methods for asserting behavior of an ongoing call.
type Expect struct {
	t           *testing.T
	c           *simulator.Call
	evts        []event.Event
	ready       <-chan bool
	readyToggle chan<- bool
	Terminated  bool
	mutex       sync.RWMutex
	nevers      []matcher
}

// New creates a new Expect wrapping an ongoing call.
// Expect provides utility methods for asserting behavior of an ongoing call.
func New(t *testing.T, c *simulator.Call) *Expect {
	buffer := make(chan event.Event, 64)
	c.Subscribe(buffer)
	readyVal, readyToggle := toggleChannel()

	th := Expect{
		t:           t,
		c:           c,
		ready:       readyVal,
		readyToggle: readyToggle,
		evts:        make([]event.Event, 0),
		nevers:      make([]matcher, 0),
	}

	go func() {
		for {
			select {
			case <-c.Caller.O:
			case evt, ok := <-buffer:
				if !ok {
					close(readyToggle)
					th.Terminated = true
					return
				}
				th.mutex.Lock()
				th.evts = append(th.evts, evt)
				th.runNevers(evt)
				th.mutex.Unlock()
				switch evt.Type() {
				case event.DisconnectType, event.InputType, event.TransferQueueType:
					readyToggle <- true
				case event.ModuleType:
					// fmt.Println(evt.(event.ModuleEvent).ModuleType)
					fallthrough
				default:
					readyToggle <- false
				}
			}
		}
	}()

	return &th
}

func (th *Expect) readEvents() (ok bool) {
	_, ok = <-th.ready
	return
}

func (th *Expect) cancelReady() {
	if th.Terminated {
		return
	}
	th.readyToggle <- false
}

func (th *Expect) runNevers(evt event.Event) {
	th.t.Helper()
	for _, m := range th.nevers {
		match, pass, got := m.match(evt)
		if !match {
			continue
		}
		if pass {
			th.t.Errorf("expected never %s. Got: '%s'", m.expected(), got)
		}
	}
}

func (th *Expect) run(m matcher, negate bool, unordered bool) {
	th.t.Helper()
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
			if negate {
				th.t.Errorf("expected not %s. Got: '%s'", m.expected(), got)
				return
			}
			if !unordered {
				th.evts = th.evts[i+1:]
			}
			return
		}
		gots = append(gots, got)
	}
	if negate {
		return
	}
	if len(gots) == 0 {
		th.t.Errorf("expected %s. Got nothing.", m.expected())
	} else {
		th.t.Errorf("expected %s. Got: \n%s", m.expected(), strings.Join(gots, "\n"))
	}
	if !ok {
		if th.c.Err != nil {
			th.t.Fatalf("call ended with error: %v", th.c.Err)
		} else {
			th.t.Fatal("call terminated unexpectedly")
		}
	}
}

// Caller simulates actions of the customer as part of the test.
func (th *Expect) Caller() CallerContext {
	return CallerContext{th.newTestContext()}
}

// Prompt offers assertions on prompts spoken by the IVR.
func (th *Expect) Prompt() PromptContext {
	return PromptContext{th.newTestContext()}
}

// Transfer offers assertions on transfers to flows and queues.
func (th *Expect) Transfer() TransferContext {
	return TransferContext{th.newTestContext()}
}

// Lambda offers assertions on lambda invocations.
func (th *Expect) Lambda() LambdaContext {
	return LambdaContext{th.newTestContext()}
}

// Attributes offers assertion on user attributes.
func (th *Expect) Attributes() AttributesContext {
	return AttributesContext{th.newTestContext()}
}

// To accepts an assertion function that will be run immediately.
// This can be use for modularising tests while maintaining the fluent interface.
func (th *Expect) To(assert func(expect *Expect)) {
	assert(th)
}

func (th *Expect) newTestContext() testContext {
	return testContext{expect: th, t: th.t}.init()
}

type testContext struct {
	expect         *Expect
	t              *testing.T
	matchers       matcherChain
	negateNext     bool
	matchNever     bool
	matchUnordered bool
}

func (tc testContext) init() testContext {
	tc.matchers = []matcher{}
	return tc
}

func (tc *testContext) addMatcher(m matcher) {
	if tc.negateNext {
		m = notMatcher{m}
		tc.negateNext = false
	}
	tc.matchers = append(tc.matchers, m)
	copy(tc.matchers[1:], tc.matchers)
	tc.matchers[0] = m
}

func (tc *testContext) run(m matcher) {
	tc.t.Helper()
	if tc.matchNever {
		tc.addMatcher(m)
		tc.expect.nevers = append(tc.expect.nevers, tc.matchers)
		return
	}
	negate := tc.negateNext
	tc.negateNext = false
	tc.addMatcher(m)
	tc.expect.run(tc.matchers, negate, tc.matchUnordered)
}

func (tc *testContext) not() {
	tc.negateNext = !tc.negateNext
}

func (tc *testContext) never() {
	tc.matchNever = true
}

func (tc *testContext) unordered() {
	tc.matchUnordered = true
}

type matcher interface {
	match(event.Event) (match bool, pass bool, got string)
	expected() string
}

type matcherChain []matcher

func (mc matcherChain) match(evt event.Event) (match bool, pass bool, got string) {
	if len(mc) == 0 {
		return
	}
	match, pass, got = mc[0].match(evt)
	if len(mc) == 1 {
		return
	}
	gots := make([]string, len(mc)-1)
	for i, m := range mc[1:] {
		m, p, g := m.match(evt)
		match = match && m
		pass = pass && p
		gots[i] = g
	}
	got = fmt.Sprintf("%s (%s)", got, strings.Join(gots, ","))
	return
}

func (mc matcherChain) expected() (exp string) {
	if len(mc) == 0 {
		return
	}
	exp = mc[0].expected()
	if len(mc) == 1 {
		return
	}
	exps := make([]string, len(mc)-1)
	for i, m := range mc[1:] {
		exps[i] = m.expected()
	}
	return fmt.Sprintf(`%s, %s`, exp, strings.Join(exps, " and "))
}

type notMatcher struct {
	matcher
}

func (nm notMatcher) match(evt event.Event) (match bool, pass bool, got string) {
	match, pass, got = nm.matcher.match(evt)
	pass = !pass
	return
}
func (nm notMatcher) expected() (exp string) {
	return fmt.Sprintf("not %s", nm.matcher.expected())
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
