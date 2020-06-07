package flowtest

import (
	"fmt"

	"github.com/edwardbrowncross/amazon-connect-simulator/event"
)

// AttributesContext is returned from Expect.Attributes()
type AttributesContext struct {
	testContext
}

// ToUpdate asserts that the contact attributes with the given key was set to the given value.
func (tc AttributesContext) ToUpdate(key string, value string) {
	tc.run(attributeKeyValueMatcher{key, value})
}

// ToUpdateKey asserts that the contact attributes with the given key was set to something.
func (tc AttributesContext) ToUpdateKey(key string) {
	tc.run(attributeKeyValueMatcher{key, "*"})
}

// Not negates the meaning of the following assertion.
func (tc AttributesContext) Not() AttributesContext {
	tc.not()
	return tc
}

// Never asserts that the following assertions will never match for the durtion of the call.
func (tc AttributesContext) Never() AttributesContext {
	tc.never()
	return tc
}

// Unordered suspends the implicit assertion that events occur in the flow in the order you assert them in your tests.
func (tc AttributesContext) Unordered() AttributesContext {
	tc.unordered()
	return tc
}

type attributeKeyValueMatcher struct {
	key   string
	value string
}

func (m attributeKeyValueMatcher) match(evt event.Event) (match bool, pass bool, got string) {
	if evt.Type() != event.UpdateContactDataType {
		return false, false, ""
	}
	e := evt.(event.UpdateContactDataEvent)
	match = true
	got = fmt.Sprintf("%s='%s'", e.Key, e.Value)
	pass = bool((e.Key == m.key || m.key == "*") && (e.Value == m.value || m.value == "*"))
	return
}

func (m attributeKeyValueMatcher) expected() string {
	key := m.key
	value := fmt.Sprintf(" to '%s'", m.value)
	if key == "*" {
		key = "any"
	}
	if value == "*" {
		value = ""
	}
	return fmt.Sprintf("to set %s field in contact data%s", key, value)
}
