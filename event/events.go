package event

import (
	"time"

	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

// Type indicates the type of event.
type Type string

// These are the types of event prevent by this package.
const (
	ModuleType            Type = "Module"
	PromptType                 = "Prompt"
	InputType                  = "Type"
	TransferQueueType          = "TransferQueue"
	TransferFlowType           = "TransferFlow"
	DisconnectType             = "Disconnect"
	UpdateContactDataType      = "UpdateContactData"
)

// Event is an event describing activity in an ongoing call.
type Event interface {
	Type() Type
}

// ModuleEvent is emitted every time a module is started.
type ModuleEvent struct {
	ID         flow.ModuleID
	ModuleType flow.ModuleType
}

// NewModuleEvent creates a ModuleEvent from data stored in a Module.
func NewModuleEvent(m flow.Module) ModuleEvent {
	return ModuleEvent{
		ID:         m.ID,
		ModuleType: m.Type,
	}
}

// Type returns PromptType.
func (e ModuleEvent) Type() Type {
	return ModuleType
}

// PromptEvent is emitted when a module outputs spoken text to the caller.
type PromptEvent struct {
	Text string
	SSML bool
}

// Type returns PromptType.
func (e PromptEvent) Type() Type {
	return PromptType
}

// InputEvent is emitted when a module is waiting for caller input.
type InputEvent struct {
	MaxDigits int
	Timeout   time.Duration
}

// Type returns InputType.
func (e InputEvent) Type() Type {
	return InputType
}

// QueueTransferEvent is emitted when a caller is transfered to a queue.
type QueueTransferEvent struct {
	QueueARN  string
	QueueName string
}

// Type returns TransferQueueType.
func (e QueueTransferEvent) Type() Type {
	return TransferQueueType
}

// FlowTransferEvent is emitted when a caller is transfered to a different flow.
type FlowTransferEvent struct {
	FlowARN  string
	FlowName string
}

// Type returns TransferFlowType.
func (e FlowTransferEvent) Type() Type {
	return TransferFlowType
}

// DisconnectEvent is emitted when the flow is terminated.
type DisconnectEvent struct{}

// Type returns DisconnectType.
func (e DisconnectEvent) Type() Type {
	return DisconnectType
}

// UpdateContactDataEvent is emitted when a field in the user data is set or updated.
type UpdateContactDataEvent struct {
	Key   string
	Value string
}

// Type returns UpdateContactDataType.
func (e UpdateContactDataEvent) Type() Type {
	return UpdateContactDataType
}
