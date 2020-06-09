package flowtest

import (
	"fmt"
	"strings"

	"github.com/edwardbrowncross/amazon-connect-simulator/event"
)

// PromptContext is returned from Expect.Prompt()
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
	tc.t.Helper()
	tc.run(promptPartialMatcher{msg})
}

// ToEqual asserts that the prompt is exacly equal to the given string.
func (tc PromptContext) ToEqual(msg string) {
	tc.t.Helper()
	tc.run(promptExactMatcher{msg})
}

// ToPlay asserts that any prompt is heard.
func (tc PromptContext) ToPlay() {
	tc.t.Helper()
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

// Unordered suspends the implicit assertion that events occur in the flow in the order you assert them in your tests.
func (tc PromptContext) Unordered() PromptContext {
	tc.unordered()
	return tc
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
