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
			key, kok := mod.Parameters.Get("key")
			val, vok := mod.Parameters.Get("value")
			if !kok || !vok {
				continue
			}
			mod.Parameters = ModuleParameterList{
				{Name: "Attribute", Key: key.Value.(string), Value: val.Value, Namespace: val.Namespace},
			}
			flow.Modules[i] = mod
		case ModuleDeprecatedTransferToFlow:
			cfid, ok := mod.Parameters.Get("ContactFlowId")
			var metadata struct {
				ContactFlow struct {
					Text string `json:"text"`
				} `json:"ContactFlow"`
			}
			err := json.Unmarshal(mod.Metadata, &metadata)
			if !ok || err != nil {
				continue
			}
			mod.Type = ModuleTransfer
			cfid.ResourceName = metadata.ContactFlow.Text
			mod.Parameters = ModuleParameterList{
				cfid,
			}
			mod.Target = TargetFlow
			flow.Modules[i] = mod
		case ModuleDeprecatedCustomerInQueue:
			mod.Type = ModuleTransfer
			mod.Target = TargetQueue
			flow.Modules[i] = mod
		case ModuleSetQueue:
			q, ok := mod.Parameters.Get("Queue")
			if mod.Target == ModuleSetQueue || ok && q.ResourceName == "" {
				var metadata struct {
					Queue struct {
						Text string `json:"text"`
					} `json:"queue"`
				}
				err := json.Unmarshal(mod.Metadata, &metadata)
				if err != nil {
					continue
				}
				q.ResourceName = metadata.Queue.Text
				mod.Parameters = ModuleParameterList{
					q,
				}
				flow.Modules[i] = mod
			}
		}
	}
	return flow
}
