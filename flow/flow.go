package flow

type Flow struct {
	Modules  []Module     `json:"modules"`
	Start    ModuleID     `json:"start"`
	Metadata FlowMetadata `json:"metadata"`
}

type FlowMetadata struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
}
