package flowtest

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/edwardbrowncross/amazon-connect-simulator/event"
)

// LambdaContext is returned from Expect.Lambda()
type LambdaContext struct {
	testContext
}

// WithTimeout adds an assertion that the timeout of the invoked lamda is equal to the given value.
func (tc LambdaContext) WithTimeout(time time.Duration) LambdaContext {
	tc.addMatcher(lambdaTimeoutMatcher{time})
	return tc
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
	tc.t.Helper()
	tc.run(lambdaCallMatcher{})
}

// ToReturnError asserts that a lambda was invoked.
func (tc LambdaContext) ToReturnError(err error) {
	tc.t.Helper()
	tc.run(lambdaErrorMatcher{err})
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

// Unordered suspends the implicit assertion that events occur in the flow in the order you assert them in your tests.
func (tc LambdaContext) Unordered() LambdaContext {
	tc.unordered()
	return tc
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
	return "to invoke lambda"
}

type lambdaErrorMatcher struct {
	err error
}

func (m lambdaErrorMatcher) match(evt event.Event) (match bool, pass bool, got string) {
	if evt.Type() != event.InvokeLambdaType {
		return false, false, ""
	}
	e := evt.(event.InvokeLambdaEvent)
	match = true
	pass = e.Error != nil
	if e.Error == nil {
		got = "nil"
	} else {
		got = e.Error.Error()
	}
	return
}

func (m lambdaErrorMatcher) expected() string {
	return fmt.Sprintf("to return error: %v", m.err)
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

type lambdaTimeoutMatcher struct {
	timeout time.Duration
}

func (m lambdaTimeoutMatcher) match(evt event.Event) (match bool, pass bool, got string) {
	if evt.Type() != event.InvokeLambdaType {
		return false, false, ""
	}
	e := evt.(event.InvokeLambdaEvent)
	match = true
	got = fmt.Sprintf("%d seconds", e.Timeout/time.Second)
	pass = e.Timeout == m.timeout
	return
}

func (m lambdaTimeoutMatcher) expected() string {
	return fmt.Sprintf("with timeout of %d seconds", m.timeout/time.Second)
}
