package simulator

type ModuleID string
type ModuleType string
type ModuleTarget string
type ModuleBranchCondition string
type ModuleParameterNamespace string
type ModuleBranchConditionType string

const (
	ModuleStoreUserInput         ModuleType = "StoreUserInput"
	ModuleCheckAttribute                    = "CheckAttribute"
	ModuleTransfer                          = "Transfer"
	ModulePlayPrompt                        = "PlayPrompt"
	ModuleDisconnect                        = "Disconnect"
	ModuleSetQueue                          = "SetQueue"
	ModuleGetUserInput                      = "GetUserInput"
	ModuleSetAttributes                     = "SetAttributes"
	ModuleInvokeExternalResource            = "InvokeExternalResource"
	ModuleCheckHoursOfOperation             = "CheckHoursOfOperation"
)

const (
	TargetFlow   ModuleTarget = "Flow"
	TargetLambda              = "Lambda"
	TargetQueue               = "Queue"
	TargetDigits              = "Digits"
)

const (
	NamespaceExternal    ModuleParameterNamespace = "External"
	NamespaceSystem                               = "System"
	NamespaceUserDefined                          = "User Defined"
)

const (
	BranchSuccess  ModuleBranchCondition = "Success"
	BranchError                          = "Error"
	BranchNoMatch                        = "NoMatch"
	BranchEvaluate                       = "Evaluate"
	BranchTimeout                        = "Timeout"
	BranchTrue                           = "True"
	BranchFalse                          = "False"
)

const (
	ConditionEquals ModuleBranchConditionType = "Equals"
	ConditionGTE                              = "GreaterThanOrEqualTo"
	ConditionGT                               = "GreaterThan"
	ConditionLTE                              = "LessThanOrEqualTo"
	ConditionLT                               = "LessThan"
)

type Flow struct {
	Modules  []Module     `json:"modules"`
	Start    ModuleID     `json:"start"`
	Metadata FlowMetadata `json:"metadata"`
}

type Module struct {
	ID         ModuleID            `json:"id"`
	Type       ModuleType          `json:"type"`
	Branches   ModuleBranchList    `json:"branches"`
	Parameters ModuleParameterList `json:"parameters"`
	Metadata   ModuleMetadata      `json:"metadata"`
	Target     ModuleTarget        `json:"target"`
}

type ModuleBranchList []ModuleBranch

func (mbl ModuleBranchList) GetLink(named ModuleBranchCondition) *ModuleID {
	for _, p := range mbl {
		if p.Condition == named {
			return &p.Transition
		}
	}
	return nil
}

func (mbl ModuleBranchList) List(named ModuleBranchCondition) []ModuleBranch {
	r := []ModuleBranch{}
	for _, b := range mbl {
		if b.Condition == named {
			r = append(r, b)
		}
	}
	return r
}

type ModuleBranch struct {
	Condition      ModuleBranchCondition     `json:"condition"`
	ConditionType  ModuleBranchConditionType `json:"conditionType"`
	ConditionValue string                    `json:"conditionValue"`
	Transition     ModuleID                  `json:"transition"`
}

type ModuleParameterList []ModuleParameter

func (mpl ModuleParameterList) Get(named string) *ModuleParameter {
	for _, p := range mpl {
		if p.Name == named {
			return &p
		}
	}
	return nil
}

func (mpl ModuleParameterList) List(named string) []ModuleParameter {
	r := []ModuleParameter{}
	for _, p := range mpl {
		if p.Name == named {
			r = append(r, p)
		}
	}
	return r
}

type ModuleParameter struct {
	Name         string                    `json:"name"`
	Key          string                    `json:"key"`
	Value        interface{}               `json:"value"`
	Namespace    *ModuleParameterNamespace `json:"namespace"`
	ResourceName string                    `json:"resourceName"`
}

type ModuleMetadata struct {
}

type FlowMetadata struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
}
