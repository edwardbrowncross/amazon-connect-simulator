package simulator

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
	"github.com/edwardbrowncross/amazon-connect-simulator/module"
)

var sampleWelcome = `{
    "modules":[
        {"id":"a456069e-9995-4119-9427-bd63308fa17f","type":"SetAttributes","branches":[{"condition":"Success","transition":"6063b277-5cd1-41fc-a069-ae76887f2a23"},{"condition":"Error","transition":"6063b277-5cd1-41fc-a069-ae76887f2a23"}],"parameters":[{"name":"Attribute","value":"true","key":"greetingPlayed","namespace":null}],"metadata":{"position":{"x":931,"y":538}}},
        {"id":"4cb86557-b541-46a3-a452-cee0b241a3cf","type":"CheckAttribute","branches":[{"condition":"Evaluate","conditionType":"Equals","conditionValue":"true","transition":"98a70ec0-069b-44a4-ada6-2a1810b1c675"},{"condition":"NoMatch","transition":"a456069e-9995-4119-9427-bd63308fa17f"}],"parameters":[{"name":"Attribute","value":"greetingPlayed"},{"name":"Namespace","value":"User Defined"}],"metadata":{"position":{"x":710,"y":362},"conditionMetadata":[{"id":"f34dc49a-9d86-419e-83e8-5c8f0c3f0fed","operator":{"name":"Equals","value":"Equals","shortDisplay":"="},"value":"true"}]}},
        {"id":"6063b277-5cd1-41fc-a069-ae76887f2a23","type":"PlayPrompt","branches":[{"condition":"Success","transition":"98a70ec0-069b-44a4-ada6-2a1810b1c675"}],"parameters":[{"name":"Text","value":"Hello, thanks for calling. These are some examples of what the Amazon Connect virtual contact center can enable you to do.","namespace":null},{"name":"TextToSpeechType","value":"text"}],"metadata":{"position":{"x":1151,"y":532},"useDynamic":false}},
        {"id":"83b76e76-52cc-4732-81ff-1519b0c0f132","type":"Disconnect","branches":[],"parameters":[],"metadata":{"position":{"x":2207,"y":607}}},
        {"id":"f9d359f6-ab58-416a-8d69-d721cf49a2df","type":"Transfer","branches":[{"condition":"Error","transition":"7c0a92c0-3a0d-4941-9d0e-2ae0e16ce58b"}],"parameters":[{"name":"ContactFlowId","value":"arn:aws:connect:eu-west-2:456789012345:instance/ffffffff-ffff-4000-ffff-ffffffffffff/contact-flow/b788fec7-2133-43ec-b6f8-5c4c98c13771","resourceName":"Sample secure input with no agent"}],"metadata":{"position":{"x":1705,"y":201},"useDynamic":false,"ContactFlow":{"id":"arn:aws:connect:eu-west-2:456789012345:instance/ffffffff-ffff-4000-ffff-ffffffffffff/contact-flow/b788fec7-2133-43ec-b6f8-5c4c98c13771","text":"Sample secure input with no agent"}},"target":"Flow"},
        {"id":"6d68b65b-787d-4359-a58b-23959a0c18d7","type":"Transfer","branches":[{"condition":"Error","transition":"7c0a92c0-3a0d-4941-9d0e-2ae0e16ce58b"}],"parameters":[{"name":"ContactFlowId","value":"arn:aws:connect:eu-west-2:456789012345:instance/ffffffff-ffff-4000-ffff-ffffffffffff/contact-flow/39b54266-74e6-416e-bd34-06e3bc6592af","resourceName":"Sample note for screenpop"}],"metadata":{"position":{"x":1703,"y":518},"useDynamic":false,"ContactFlow":{"id":"arn:aws:connect:eu-west-2:456789012345:instance/ffffffff-ffff-4000-ffff-ffffffffffff/contact-flow/39b54266-74e6-416e-bd34-06e3bc6592af","text":"Sample note for screenpop"}},"target":"Flow"},
        {"id":"4c568499-9a87-46d1-87e2-213ebaf81c4e","type":"Transfer","branches":[{"condition":"Error","transition":"7c0a92c0-3a0d-4941-9d0e-2ae0e16ce58b"}],"parameters":[{"name":"ContactFlowId","value":"arn:aws:connect:eu-west-2:456789012345:instance/ffffffff-ffff-4000-ffff-ffffffffffff/contact-flow/d62e442f-eb29-437c-a33c-9f29664b229c","resourceName":"Sample Lambda integration"}],"metadata":{"position":{"x":1706,"y":364},"useDynamic":false,"ContactFlow":{"id":"arn:aws:connect:eu-west-2:456789012345:instance/ffffffff-ffff-4000-ffff-ffffffffffff/contact-flow/d62e442f-eb29-437c-a33c-9f29664b229c","text":"Sample Lambda integration"}},"target":"Flow"},
        {"id":"3b843a8a-19c1-4f34-b874-c28a8550a352","type":"Transfer","branches":[{"condition":"Error","transition":"7c0a92c0-3a0d-4941-9d0e-2ae0e16ce58b"}],"parameters":[{"name":"ContactFlowId","value":"arn:aws:connect:eu-west-2:456789012345:instance/ffffffff-ffff-4000-ffff-ffffffffffff/contact-flow/25625805-4c2f-4c09-8a9b-b9008fdc0eee","resourceName":"Sample recording behavior"}],"metadata":{"position":{"x":1703,"y":833},"useDynamic":false,"ContactFlow":{"id":"arn:aws:connect:eu-west-2:456789012345:instance/ffffffff-ffff-4000-ffff-ffffffffffff/contact-flow/25625805-4c2f-4c09-8a9b-b9008fdc0eee","text":"Sample recording behavior"}},"target":"Flow"},
        {"id":"b11b10a6-c4bd-41ba-a9d9-098cc5374035","type":"CheckAttribute","branches":[{"condition":"Evaluate","conditionType":"Equals","conditionValue":"CHAT","transition":"fe6810db-1fe8-4b8d-a939-65dadbb61ef7"},{"condition":"NoMatch","transition":"4cb86557-b541-46a3-a452-cee0b241a3cf"}],"parameters":[{"name":"Attribute","value":"Channel"},{"name":"Namespace","value":"System"}],"metadata":{"position":{"x":491,"y":42},"conditionMetadata":[{"id":"3333aad8-3a27-4632-bfb8-46e8e326d783","operator":{"name":"Equals","value":"Equals","shortDisplay":"="},"value":"CHAT"}]}},
        {"id":"6e53ccd1-94bf-4f4e-a6c7-69d1f2b9cc20","type":"SetEventHook","branches":[{"condition":"Success","transition":"b11b10a6-c4bd-41ba-a9d9-098cc5374035"},{"condition":"Error","transition":"b11b10a6-c4bd-41ba-a9d9-098cc5374035"}],"parameters":[{"name":"Type","value":"CustomerRemaining"},{"name":"ContactFlowId","value":"arn:aws:connect:eu-west-2:456789012345:instance/ffffffff-ffff-4000-ffff-ffffffffffff/contact-flow/2e01a339-1f61-443d-a3d2-c65eb0fa053e","resourceName":"Sample disconnect flow"}],"metadata":{"position":{"x":207,"y":43},"useDynamic":false,"ContactFlow":{"id":"arn:aws:connect:eu-west-2:456789012345:instance/ffffffff-ffff-4000-ffff-ffffffffffff/contact-flow/2e01a339-1f61-443d-a3d2-c65eb0fa053e","text":"Sample disconnect flow"}},"target":"Flow"},
        {"id":"fe6810db-1fe8-4b8d-a939-65dadbb61ef7","type":"PlayPrompt","branches":[{"condition":"Success","transition":"ae54d6c1-507d-4d6e-886e-e212cbbcf976"}],"parameters":[{"name":"Text","value":"Hello, thanks for contacting us. This is an example of what the Amazon Connect virtual contact center can enable you to do.","namespace":null},{"name":"TextToSpeechType","value":"text"}],"metadata":{"position":{"x":1379,"y":44},"useDynamic":false}},
        {"id":"ae54d6c1-507d-4d6e-886e-e212cbbcf976","type":"Transfer","branches":[{"condition":"Error","transition":"7c0a92c0-3a0d-4941-9d0e-2ae0e16ce58b"}],"parameters":[{"name":"ContactFlowId","value":"arn:aws:connect:eu-west-2:456789012345:instance/ffffffff-ffff-4000-ffff-ffffffffffff/contact-flow/c099863a-8809-4276-9f63-8d39fed2ba31","resourceName":"Sample Queue Configurations Flow"}],"metadata":{"position":{"x":1701,"y":41},"useDynamic":false,"ContactFlow":{"id":"arn:aws:connect:eu-west-2:456789012345:instance/ffffffff-ffff-4000-ffff-ffffffffffff/contact-flow/c099863a-8809-4276-9f63-8d39fed2ba31","text":"Sample Queue Configurations Flow"}},"target":"Flow"},
        {"id":"7c0a92c0-3a0d-4941-9d0e-2ae0e16ce58b","type":"PlayPrompt","branches":[{"condition":"Success","transition":"83b76e76-52cc-4732-81ff-1519b0c0f132"}],"parameters":[{"name":"Text","value":"We're sorry, an error occurred. Goodbye.","namespace":null},{"name":"TextToSpeechType","value":"text"}],"metadata":{"position":{"x":1984,"y":606},"useDynamic":false}},
        {"id":"98a70ec0-069b-44a4-ada6-2a1810b1c675","type":"GetUserInput","branches":[{"condition":"Evaluate","conditionType":"Equals","conditionValue":"1","transition":"ae54d6c1-507d-4d6e-886e-e212cbbcf976"},{"condition":"Evaluate","conditionType":"Equals","conditionValue":"2","transition":"f9d359f6-ab58-416a-8d69-d721cf49a2df"},{"condition":"Evaluate","conditionType":"Equals","conditionValue":"3","transition":"4c568499-9a87-46d1-87e2-213ebaf81c4e"},{"condition":"Evaluate","conditionType":"Equals","conditionValue":"4","transition":"6d68b65b-787d-4359-a58b-23959a0c18d7"},{"condition":"Evaluate","conditionType":"Equals","conditionValue":"5","transition":"f7ee6062-55b6-4bcf-9b06-35e5735fbad7"},{"condition":"Evaluate","conditionType":"Equals","conditionValue":"6","transition":"3b843a8a-19c1-4f34-b874-c28a8550a352"},{"condition":"Timeout","transition":"ae54d6c1-507d-4d6e-886e-e212cbbcf976"},{"condition":"NoMatch","transition":"3b843a8a-19c1-4f34-b874-c28a8550a352"},{"condition":"Error","transition":"7c0a92c0-3a0d-4941-9d0e-2ae0e16ce58b"}],"parameters":[{"name":"Text","value":"Press 1 to be put in queue for an agent.\n2 to securely enter content. \n3 to hear the results of an AWS Lambda data dip. \n4 to set a screen pop for the agent. \n5 to roll the dice and simulate a and b testing. \nOr 6 to set call recording behavior.","namespace":null},{"name":"TextToSpeechType","value":"text"},{"name":"Timeout","value":"8"},{"name":"MaxDigits","value":"1"}],"metadata":{"position":{"x":1383,"y":360},"conditionMetadata":[{"id":"14d34c04-a158-433b-924c-e8a4c3938895","value":"1"},{"id":"f74b4db5-77f9-4b40-961a-7e280f7b6420","value":"2"},{"id":"4a8cf615-b476-44c5-adcf-50c130504767","value":"3"},{"id":"b56c72e8-21e4-484f-8b59-ed161f798594","value":"4"},{"id":"325435f4-db3a-46ab-89d2-c1c44903ef13","value":"5"},{"id":"e81533fc-be84-4206-b5cb-eba9984bd8ff","value":"6"}],"useDynamic":false},"target":"Digits"},
        {"id":"f7ee6062-55b6-4bcf-9b06-35e5735fbad7","type":"Transfer","branches":[{"condition":"Error","transition":"7c0a92c0-3a0d-4941-9d0e-2ae0e16ce58b"}],"parameters":[{"name":"ContactFlowId","value":"arn:aws:connect:eu-west-2:456789012345:instance/ffffffff-ffff-4000-ffff-ffffffffffff/contact-flow/9e76597b-b690-4bbb-85eb-1c031e916816","resourceName":"Sample AB test"}],"metadata":{"position":{"x":1705,"y":675},"useDynamic":false,"ContactFlow":{"id":"arn:aws:connect:eu-west-2:456789012345:instance/ffffffff-ffff-4000-ffff-ffffffffffff/contact-flow/9e76597b-b690-4bbb-85eb-1c031e916816","text":"Sample AB test","arn":null,"metricDetail":null,"resourceId":null}},"target":"Flow"}
    ],
    "version":"1","type":"contactFlow","start":"6e53ccd1-94bf-4f4e-a6c7-69d1f2b9cc20",
    "metadata":{"entryPointPosition":{"x":20,"y":41},"snapToGrid":false,"name":"Sample inbound flow (first contact experience)","description":"First contact experience","type":"contactFlow","status":"published","hash":"9a771dbb182b19ce2522d2c19e55bee3e92489e6093e43283133a276c7b90eda"}
}`

var sampleLambda = `{
    "modules":[
        {"id":"7329da0c-3dcb-4661-a72e-95b6e841a4a4","type":"PlayPrompt","branches":[{"condition":"Success","transition":"1b9a1e90-a330-450b-85a9-dcad8ef3b045"}],"parameters":[{"name":"Text","value":"Based on the number you are calling from, your area code is located in $.External.State","namespace":null},{"name":"TextToSpeechType","value":"text"}],"metadata":{"position":{"x":928,"y":187},"useDynamic":false}},
        {"id":"96f62f74-1905-40cd-acca-714c0782717a","type":"Disconnect","branches":[],"parameters":[],"metadata":{"position":{"x":1670,"y":475}}},
        {"id":"94cd8c74-9a86-41bd-8fe2-d08bc8f9e41e","type":"Transfer","branches":[{"condition":"Error","transition":"31dd3a3e-7d66-4829-9252-8ea344160f5e"}],"parameters":[{"name":"ContactFlowId","value":"arn:aws:connect:eu-west-2:456789012345:instance/ffffffff-ffff-4000-ffff-ffffffffffff/contact-flow/f047d1a0-c42c-4660-9e9f-fa684b7d4493","resourceName":"Sample inbound flow (first contact experience)"}],"metadata":{"position":{"x":1195,"y":314},"useDynamic":false,"ContactFlow":{"id":"arn:aws:connect:eu-west-2:456789012345:instance/ffffffff-ffff-4000-ffff-ffffffffffff/contact-flow/f047d1a0-c42c-4660-9e9f-fa684b7d4493","text":"Sample inbound flow (first contact experience)"}},"target":"Flow"},
        {"id":"31dd3a3e-7d66-4829-9252-8ea344160f5e","type":"PlayPrompt","branches":[{"condition":"Success","transition":"96f62f74-1905-40cd-acca-714c0782717a"}],"parameters":[{"name":"Text","value":"Failed to transfer back to main flow","namespace":null},{"name":"TextToSpeechType","value":"text"}],"metadata":{"position":{"x":1425,"y":392},"useDynamic":false}},
        {"id":"35c77601-311e-4e0b-85a5-883381ac2655","type":"CheckAttribute","branches":[{"condition":"Evaluate","conditionType":"Equals","conditionValue":"unknown","transition":"431f29e2-cca7-44e4-a449-90a38c2d327b"},{"condition":"NoMatch","transition":"7329da0c-3dcb-4661-a72e-95b6e841a4a4"}],"parameters":[{"name":"Attribute","value":"State"},{"name":"Namespace","value":"External"}],"metadata":{"position":{"x":672,"y":212},"conditionMetadata":[{"id":"6ce7e2b8-e08e-4dc0-bddf-4e7eadd55951","operator":{"name":"Equals","value":"Equals","shortDisplay":"="},"value":"unknown"}]}},
        {"id":"1b9a1e90-a330-450b-85a9-dcad8ef3b045","type":"PlayPrompt","branches":[{"condition":"Success","transition":"94cd8c74-9a86-41bd-8fe2-d08bc8f9e41e"}],"parameters":[{"name":"Text","value":"Now returning you to the main menu.","namespace":null},{"name":"TextToSpeechType","value":"text"}],"metadata":{"position":{"x":1181,"y":88},"useDynamic":false}},
        {"id":"5d737fb6-6df3-4e27-beff-eb3395bada65","type":"PlayPrompt","branches":[{"condition":"Success","transition":"1b9a1e90-a330-450b-85a9-dcad8ef3b045"}],"parameters":[{"name":"Text","value":"Here is your fun fact: $.External.Fact","namespace":null},{"name":"TextToSpeechType","value":"text"}],"metadata":{"position":{"x":685,"y":37},"useDynamic":false}},
        {"id":"68f1b094-8c1c-4231-879d-b106e53de281","type":"CheckAttribute","branches":[{"condition":"Evaluate","conditionType":"Equals","conditionValue":"CHAT","transition":"5d737fb6-6df3-4e27-beff-eb3395bada65"},{"condition":"NoMatch","transition":"35c77601-311e-4e0b-85a5-883381ac2655"}],"parameters":[{"name":"Attribute","value":"Channel"},{"name":"Namespace","value":"System"}],"metadata":{"position":{"x":439,"y":121},"conditionMetadata":[{"id":"b687a803-0487-43a4-8bbb-4acff874039c","operator":{"name":"Equals","value":"Equals","shortDisplay":"="},"value":"CHAT"}]}},
        {"id":"431f29e2-cca7-44e4-a449-90a38c2d327b","type":"PlayPrompt","branches":[{"condition":"Success","transition":"1b9a1e90-a330-450b-85a9-dcad8ef3b045"}],"parameters":[{"name":"Text","value":"Sorry, we failed to find the state for your phone number's area code.","namespace":null},{"name":"TextToSpeechType","value":"text"}],"metadata":{"position":{"x":917,"y":387},"useDynamic":false}},
        {"id":"4d36a741-bc87-4035-b3fa-9c8390e687ac","type":"PlayPrompt","branches":[{"condition":"Success","transition":"7eefafd6-402f-4759-967c-b017ef5f3969"}],"parameters":[{"name":"Text","value":"Now performing a data dip using AWS Lambda. Based on your phone number, we will lookup the state you are calling from if you are on a voice call or tell you a fun fact if you are on chat.","namespace":null},{"name":"TextToSpeechType","value":"text"}],"metadata":{"position":{"x":193,"y":50},"useDynamic":false}},
        {"id":"7eefafd6-402f-4759-967c-b017ef5f3969","type":"InvokeExternalResource","branches":[{"condition":"Success","transition":"68f1b094-8c1c-4231-879d-b106e53de281"},{"condition":"Error","transition":"431f29e2-cca7-44e4-a449-90a38c2d327b"}],"parameters":[{"name":"FunctionArn","value":"arn:aws:lambda:us-east-1:613787477748:function:state-lookup","namespace":null},{"name":"TimeLimit","value":"4"}],"metadata":{"position":{"x":150,"y":245},"dynamicMetadata":{},"useDynamic":false},"target":"Lambda"}
    ],
    "version":"1","type":"contactFlow","start":"4d36a741-bc87-4035-b3fa-9c8390e687ac",
    "metadata":{"entryPointPosition":{"x":39,"y":15},"snapToGrid":false,"name":"Sample Lambda integration","description":"Invokes a lambda function to determine information about the user.","type":"contactFlow","status":"published","hash":"9345d0d8ef8e65010f8d036c17560c9fc47bf60146e8f6ed5ff26b5cdd1d5fc3"}
}`

func TestSimulator(t *testing.T) {
	// Create a simulator.
	sim := New()

	// Load flow struct.
	f := flow.Flow{}
	json.Unmarshal([]byte(sampleWelcome), &f)
	sim.LoadFlow(f)

	// Load bad json. Expect error.
	var err error
	err = sim.LoadFlowJSON([]byte(`<xml type="flow" />`))
	if err == nil {
		t.Error("expected error loading xml, but got no error")
	}
	err = sim.LoadFlowJSON([]byte(`[{"type":"PlayPrompt"}]`))
	if err == nil {
		t.Error("expected error loading json array, but got no error")
	}

	// Load good json.
	err = sim.LoadFlowJSON([]byte(sampleLambda))
	if err != nil {
		t.Fatalf("unexpected error parsing sample lambda: %v", err)
	}

	type stateLookupOutput struct {
		State string `json:"State"`
	}

	// Register a lambda.
	err = sim.RegisterLambda("state-lookup", func(ctx context.Context, in module.LambdaPayload) (out stateLookupOutput, err error) {
		return stateLookupOutput{
			State: "United Kingdom",
		}, nil
	})
	if err != nil {
		t.Fatalf("unexpected error registering lambda: %v", err)
	}

	// Register a lambda with the wrong signature. Expect error.
	err = sim.RegisterLambda("state-lookup", func(string) error { return nil })
	if err == nil {
		t.Error("expected error registering incompatible lambda handler but got none.")
	}

	var call Call

	// Try to start call without starting flow set. Expect error.
	_, err = sim.StartCall(CallConfig{})
	if err == nil {
		t.Error("expected error starting call early but got none.")
	}

	// Set incorrect starting flow. Expect error.
	err = sim.SetStartingFlow("Sample self destruct flow")
	if err == nil {
		t.Error("expected error setting non-existant starting flow but got none.")
	}

	// Set correct starting flow.
	err = sim.SetStartingFlow("Sample inbound flow (first contact experience)")
	if err != nil {
		t.Fatalf("unexpected error setting starting flow: %v", err)
	}

	// Start a call.
	call, err = sim.StartCall(CallConfig{
		SourceNumber: "+447878123456",
		DestNumber:   "+441121234567",
	})
	if err != nil {
		t.Fatalf("unexpected error starting call: %v", err)
	}
	var prompt string
	var expPrompt string
	prompt = <-call.O
	expPrompt = "Hello, thanks for calling. These are some examples of what the Amazon Connect virtual contact center can enable you to do."
	if prompt != expPrompt {
		t.Errorf("Expected welcome prompt of \n'%s'\n but got \n'%s'", expPrompt, prompt)
	}
}
