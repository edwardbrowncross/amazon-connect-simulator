package flow

import (
	"encoding/json"
)

// Dedeprecate removes deprecated modules from a flow and replaces them with the current equivalent.
func Dedeprecate(flow Flow) Flow {
	for i, mod := range flow.Modules {
		switch mod.Type {
		case ModuleDeprecatedPlayAudio:
			mod.Type = ModulePlayPrompt
			flow.Modules[i] = mod
		case ModuleDeprecatedStoreCustomerInput:
			mod.Type = ModuleStoreUserInput
			flow.Modules[i] = mod
		case ModuleDeprecatedSetScreenPop:
			mod.Type = ModuleSetAttributes
			key := mod.Parameters.Get("key")
			val := mod.Parameters.Get("value")
			if key == nil || val == nil {
				continue
			}
			mod.Parameters = ModuleParameterList{
				{Name: "Attribute", Key: key.Value.(string), Value: val.Value, Namespace: val.Namespace},
			}
			flow.Modules[i] = mod
		case ModuleDeprecatedTransferToFlow:
			cfid := mod.Parameters.Get("ContactFlowId")
			var metadata struct {
				ContactFlow struct {
					Text string `json:"text"`
				} `json:"ContactFlow"`
			}
			err := json.Unmarshal(mod.Metadata, &metadata)
			if cfid == nil || err != nil {
				continue
			}
			mod.Type = ModuleTransfer
			cfid.ResourceName = metadata.ContactFlow.Text
			mod.Parameters = ModuleParameterList{
				*cfid,
			}
			mod.Target = TargetFlow
			flow.Modules[i] = mod
		case ModuleDeprecatedCustomerInQueue:
			mod.Type = ModuleTransfer
			mod.Target = TargetQueue
			flow.Modules[i] = mod
		case ModuleSetQueue:
			if mod.Target == ModuleSetQueue || mod.Parameters.Get("Queue") != nil && mod.Parameters.Get("Queue").ResourceName == "" {
				var metadata struct {
					Queue struct {
						Text string `json:"text"`
					} `json:"queue"`
				}
				err := json.Unmarshal(mod.Metadata, &metadata)
				if err != nil {
					continue
				}
				q := mod.Parameters.Get("Queue")
				q.ResourceName = metadata.Queue.Text
				mod.Parameters = ModuleParameterList{
					*q,
				}
				flow.Modules[i] = mod
			}
		}
	}
	return flow
}
