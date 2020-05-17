package event

// Type indicates the type of event.
type Type string

// These are the types of event prevent by this package.
const (
	PromptType Type = "Prompt"
)

// Event is an event describing activity in an ongoing call.
type Event interface {
	Type() Type
}

// PromptEvent is emitted when a module outputs spoken text to the caller.
type PromptEvent struct {
	Text string
	SSML bool
}

// Type returns the type of the event.
func (e PromptEvent) Type() Type {
	return PromptType
}
