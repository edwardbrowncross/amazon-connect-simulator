package connect

import "encoding/json"

// Payload represents the data passed to lambda by Connect when the lambdas are invoked.
type Payload struct {
	// The real information passed to the lambda.
	Details PayloadDetails `json:"Details"`
	Name    string         `json:"Name"`
}

// PayloadDetails holds data inside the connect payload.
type PayloadDetails struct {
	// Data around the customer call that this invocation is supporting.
	ContactData PayloadContactData `json:"ContactData"`
	// Values explicitly passed to the lambda from the Connect GUI.
	// It can be unmarshalled separately by the receiving lambda.
	Parameters json.RawMessage `json:"Parameters"`
}

// PayloadContactData holds information around the customer call that a lambda invocation is supporting.
type PayloadContactData struct {
	// Attributes holds variable user attributes.
	// It can be unmarshalled separately by the receiving lambda.
	Attributes        json.RawMessage `json:"Attributes"`
	ContactID         string          `json:"ContactId"`
	InitialContactID  string          `json:"InitialContactId"`
	PreviousContactID string          `json:"PreviousContactId"`
	// VOICE | CHAT
	Channel string `json:"Channel"`
	// INBOUND | OUTBOUND | TRANSFER | CALLBACK.
	InitiationMethod string `json:"InitiationMethod"`
	// What the customer is calling from.
	CustomerEndpoint PayloadContactEndpoint `json:"CustomerEndpoint"`
	// What the customer connected to to initiate their journey.
	SystemEndpoint PayloadContactEndpoint `json:"SystemEndpoint"`
	// The Connect instance that called this lambda.
	InstanceARN string `json:"InstanceARN"`
	// The Queue the customer is currently assigned to.
	Queue interface{} `json:"Queue"`
}

// PayloadContactEndpoint holds information on where the customer is calling from or connected to.
type PayloadContactEndpoint struct {
	// Phone number in a voice call.
	Address string `json:"Address"`
	// Set to TELEPHONE_NUMBER if a voice call.
	Type string `json:"Type"`
}
