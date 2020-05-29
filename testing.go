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
	nevers      []matcher
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
		nevers:      make([]matcher, 0),
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
				th.runNevers(evt)
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

func (th *TestHelper) runNevers(evt event.Event) {
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

func (th *TestHelper) run(m matcher, negate bool) {
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
			th.evts = th.evts[i+1:]
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

// Prompt offers assertions on prompts spoken by the IVR.
func (th *TestHelper) Prompt() PromptContext {
	return PromptContext{th.newTestContext()}
}

// Transfer offers assertions on transfers to flows and queues.
func (th *TestHelper) Transfer() TransferContext {
	return TransferContext{th.newTestContext()}
}

// Lambda offers assertions on lambda invocations.
func (th *TestHelper) Lambda() LambdaContext {
	return LambdaContext{th.newTestContext()}
}

func (th *TestHelper) newTestContext() testContext {
	return testContext{TestHelper: th}.init()
}

type testContext struct {
	*TestHelper
	matchers   matcherChain
	negateNext bool
	matchNever bool
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
	if tc.matchNever {
		tc.addMatcher(m)
		tc.TestHelper.nevers = append(tc.TestHelper.nevers, tc.matchers)
		return
	}
	negate := tc.negateNext
	tc.negateNext = false
	tc.addMatcher(m)
	tc.TestHelper.run(tc.matchers, negate)
}

func (tc *testContext) not() {
	tc.negateNext = !tc.negateNext
}

func (tc *testContext) never() {
	tc.matchNever = true
}

// PromptContext is returned from TestHelper.Prompt()
type PromptContext struct {
	testContext
}

// WithSSML adds a pending assertion that the matching prompt also is interpreted as SSML.
func (tc PromptContext) WithSSML() PromptContext {
	tc.addMatcher(promptSSMLMatcher{true})
	return tc
}

// WithPlaintext adds a pending assertion that the matching prompt is also not interpreted as SSML.
func (tc PromptContext) WithPlaintext() PromptContext {
	tc.addMatcher(promptSSMLMatcher{false})
	return tc
}

// WithVoice adds a pending assertion that the matching prompt will be spoken with the given voice (e.g. "Joanna").
func (tc PromptContext) WithVoice(voice string) PromptContext {
	tc.addMatcher(promptVoiceMatcher{voice})
	return tc
}

// ToContain asserts that the prompt contains the given string.
func (tc PromptContext) ToContain(msg string) {
	tc.run(promptPartialMatcher{msg})
}

// ToEqual asserts that the prompt is exacly equal to the given string.
func (tc PromptContext) ToEqual(msg string) {
	tc.run(promptExactMatcher{msg})
}

// ToPlay asserts that any prompt is heard.
func (tc PromptContext) ToPlay() {
	tc.run(promptPartialMatcher{""})
}

// Not negates the meaning of the following assertion.
func (tc PromptContext) Not() PromptContext {
	tc.not()
	return tc
}

// Never asserts that the following assertions will never match for the durtion of the call.
func (tc PromptContext) Never() PromptContext {
	tc.never()
	return tc
}

// TransferContext is returned from TestHelper.Transfer()
type TransferContext struct {
	testContext
}

// ToQueue asserts that the caller is transferred to a queue with the given name.
func (tc TransferContext) ToQueue(named string) {
	tc.run(queueTransferMatcher{named})
}

// ToFlow asserts that the call moved to the flow with the given name.
func (tc TransferContext) ToFlow(named string) {
	tc.run(flowTransferMatcher{named})
}

// Never asserts that the following assertions will never match for the durtion of the call.
func (tc TransferContext) Never() TransferContext {
	tc.never()
	return tc
}

// UserAttributeUpdate asserts that a key, value pair was set in the user attributes.
func (th *TestHelper) UserAttributeUpdate(key string, value string) {
	th.run(updateContactDataMatcher{key, value}, false)
}

// LambdaContext is returned from TestHelper.Lambda()
type LambdaContext struct {
	testContext
}

// WithARN adds an assertion that the ARN of the invoked lambda also contains the given string.
func (tc LambdaContext) WithARN(arn string) LambdaContext {
	tc.addMatcher(lambdaARNMatcher{arn})
	return tc
}

// WithParameters adds an assertion that the ARN of the invoked lambda was passed custom parameters including those given.
func (tc LambdaContext) WithParameters(params map[string]string) LambdaContext {
	tc.addMatcher(lambdaParametersMatcher{params})
	return tc
}

// WithParameter adds an assertion that the ARN of the invoked lambda was passed custom parameters including the one given.
func (tc LambdaContext) WithParameter(key string, value string) LambdaContext {
	tc.addMatcher(lambdaParametersMatcher{map[string]string{key: value}})
	return tc
}

// ToBeInvoked asserts that a lambda was invoked.
func (tc LambdaContext) ToBeInvoked() {
	tc.run(lambdaCallMatcher{})
}

// Not negates the meaning of the following assertion.
func (tc LambdaContext) Not() LambdaContext {
	tc.not()
	return tc
}

// Never asserts that the following assertions will never match for the durtion of the call.
func (tc LambdaContext) Never() LambdaContext {
	tc.never()
	return tc
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
	if m.text == "" {
		return "to get a prompt"
	}
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
		got = "as SSML"
	} else {
		got = "as plaintext"
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

type promptVoiceMatcher struct {
	voice string
}

func (m promptVoiceMatcher) match(evt event.Event) (match bool, pass bool, got string) {
	if evt.Type() != event.PromptType {
		return false, false, ""
	}
	e := evt.(event.PromptEvent)
	match = true
	got = fmt.Sprintf("in %s voice", e.Voice)
	pass = bool(m.voice == e.Voice)
	return
}

func (m promptVoiceMatcher) expected() string {
	return fmt.Sprintf("read in the %s voice", m.voice)
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
	got = "invocation"
	return
}

func (m lambdaCallMatcher) expected() string {
	return fmt.Sprintf("to invoke lambda")
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
	return fmt.Sprintf("with ARN containing '%s'", m.arn)
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
