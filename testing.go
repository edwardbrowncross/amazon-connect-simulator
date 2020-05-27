package simulator

import (
	"encoding/json"
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
					fmt.Println(evt.(event.ModuleEvent).ModuleType)
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
			th.evts = th.evts[i+1:]
			return
		}
		gots = append(gots, got)
	}
	if len(gots) == 0 {
		th.t.Errorf("expected %s. Got nothing.", m.expected())
	} else {
		th.t.Errorf("expected %s. Got: \n%v", m.expected(), strings.Join(gots, "\n"))
	}
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

func (th *TestHelper) Prompt() promptContext {
	return promptContext{th.newTestContext()}
}

func (th *TestHelper) Transfer() transferContext {
	return transferContext{th.newTestContext()}
}

func (th *TestHelper) Lambda() lambdaContext {
	return lambdaContext{th.newTestContext()}
}

func (th *TestHelper) newTestContext() testContext {
	return testContext{TestHelper: th}.init()
}

type promptContext struct {
	testContext
}

func (tc promptContext) WithSSML() promptContext {
	tc.addMatcher(promptSSMLMatcher{true})
	return tc
}

func (tc promptContext) WithPlaintext() promptContext {
	tc.addMatcher(promptSSMLMatcher{false})
	return tc
}

func (tc promptContext) ToContain(msg string) {
	tc.run(promptPartialMatcher{msg})
}

func (tc promptContext) ToEqual(msg string) {
	tc.run(promptExactMatcher{msg})
}

type transferContext struct {
	testContext
}

// ToQueue asserts that the caller is transferred to a queue with the given name.
func (tc transferContext) ToQueue(named string) {
	tc.run(queueTransferMatcher{named})
}

// ToFlow asserts that the call moved to the flow with the given name.
func (tc transferContext) ToFlow(named string) {
	tc.run(flowTransferMatcher{named})
}

// UserAttributeUpdate asserts that a key, value pair was set in the user attributes.
func (th *TestHelper) UserAttributeUpdate(key string, value string) {
	th.run(updateContactDataMatcher{key, value})
}

type lambdaContext struct {
	testContext
}

func (tc lambdaContext) WithARN(arn string) lambdaContext {
	tc.addMatcher(lambdaARNMatcher{arn})
	return tc
}

func (tc lambdaContext) WithParameters(params map[string]string) lambdaContext {
	tc.addMatcher(lambdaParametersMatcher{params})
	return tc
}

// ToBeInvoked asserts that a lambda was invoked.
func (tc lambdaContext) ToBeInvoked() {
	tc.run(lambdaCallMatcher{})
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

type testContext struct {
	*TestHelper
	matchers matcherChain
}

func (tc testContext) init() testContext {
	tc.matchers = []matcher{}
	return tc
}

func (tc *testContext) addMatcher(m matcher) {
	tc.matchers = append(tc.matchers, m)
	copy(tc.matchers[1:], tc.matchers)
	tc.matchers[0] = m
}

func (tc *testContext) run(m matcher) {
	tc.addMatcher(m)
	tc.TestHelper.run(tc.matchers)
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

type promptSSMLMatcher struct {
	ssml bool
}

func (m promptSSMLMatcher) match(evt event.Event) (match bool, pass bool, got string) {
	if evt.Type() != event.PromptType {
		return false, false, ""
	}
	e := evt.(event.PromptEvent)
	match = true
	if e.SSML {
		got = "SSML"
	} else {
		got = "plaintext"
	}
	pass = bool(m.ssml == e.SSML)
	return
}

func (m promptSSMLMatcher) expected() string {
	if m.ssml {
		return "read as SSML"
	}
	return "read as plaintext"
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

type updateContactDataMatcher struct {
	key   string
	value string
}

func (m updateContactDataMatcher) match(evt event.Event) (match bool, pass bool, got string) {
	if evt.Type() != event.UpdateContactDataType {
		return false, false, ""
	}
	e := evt.(event.UpdateContactDataEvent)
	match = true
	got = fmt.Sprintf("%s='%s'", e.Key, e.Value)
	pass = bool(e.Key == m.key && e.Value == m.value)
	return
}

func (m updateContactDataMatcher) expected() string {
	return fmt.Sprintf("to set %s field in contact data to '%s'", m.key, m.value)
}

type lambdaCallMatcher struct{}

func (m lambdaCallMatcher) match(evt event.Event) (match bool, pass bool, got string) {
	if evt.Type() != event.InvokeLambdaType {
		return false, false, ""
	}
	match = true
	pass = true
	return
}

func (m lambdaCallMatcher) expected() string {
	return fmt.Sprintf("expected lambda to be invoked")
}

type lambdaARNMatcher struct {
	arn string
}

func (m lambdaARNMatcher) match(evt event.Event) (match bool, pass bool, got string) {
	if evt.Type() != event.InvokeLambdaType {
		return false, false, ""
	}
	e := evt.(event.InvokeLambdaEvent)
	match = true
	got = e.ARN
	pass = strings.Contains(e.ARN, m.arn)
	return
}

func (m lambdaARNMatcher) expected() string {
	return fmt.Sprintf("with ARN '%s'", m.arn)
}

type lambdaParametersMatcher struct {
	params map[string]string
}

func (m lambdaParametersMatcher) match(evt event.Event) (match bool, pass bool, got string) {
	if evt.Type() != event.InvokeLambdaType {
		return false, false, ""
	}
	e := evt.(event.InvokeLambdaEvent)
	match = true
	got = e.ParamsJSON
	var p map[string]string
	if err := json.Unmarshal([]byte(e.ParamsJSON), &p); err != nil {
		return
	}
	pass = true
	for k, expV := range m.params {
		if gotV, ok := p[k]; !ok || gotV != expV {
			pass = false
		}
	}
	return
}

func (m lambdaParametersMatcher) expected() string {
	return fmt.Sprintf("with parameters %v", m.params)
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
