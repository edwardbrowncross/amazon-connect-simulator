package simulator

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
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

var sampleRecording = `{
    "modules":[
        {"id":"e1cc799f-0710-42f3-a656-9772f0915925","type":"Disconnect","branches":[],"parameters":[],"metadata":{"position":{"x":1745,"y":411}}},
        {"id":"7cbaf36b-2899-43f5-b834-333a68d2067a","type":"PlayPrompt","branches":[{"condition":"Success","transition":"e1cc799f-0710-42f3-a656-9772f0915925"}],"parameters":[{"name":"Text","value":"Failed to transfer back to main phone tree","namespace":null},{"name":"TextToSpeechType","value":"text"}],"metadata":{"position":{"x":1518,"y":411},"useDynamic":false}},
        {"id":"53aab24b-bbe5-4014-b189-81f7a66a3997","type":"SetRecordingBehavior","branches":[{"condition":"Success","transition":"9b30963a-505a-425a-8a5a-c9fc2e4abd7e"}],"parameters":[{"name":"RecordingBehaviorOption","value":"Enable"},{"name":"RecordingParticipantOption","value":"Agent"}],"metadata":{"position":{"x":919,"y":196}}},
        {"id":"c19afb78-cec9-48e5-b696-22d64f2832f1","type":"SetRecordingBehavior","branches":[{"condition":"Success","transition":"9b30963a-505a-425a-8a5a-c9fc2e4abd7e"}],"parameters":[{"name":"RecordingBehaviorOption","value":"Enable"},{"name":"RecordingParticipantOption","value":"Customer"}],"metadata":{"position":{"x":921,"y":349}}},
        {"id":"e9b9dea3-74d8-4cf5-ac46-fe11369e7037","type":"SetRecordingBehavior","branches":[{"condition":"Success","transition":"9b30963a-505a-425a-8a5a-c9fc2e4abd7e"}],"parameters":[{"name":"RecordingBehaviorOption","value":"Disable"},{"name":"RecordingParticipantOption","value":"Both"}],"metadata":{"position":{"x":921,"y":500}}},
        {"id":"5a3c5b08-e33b-4485-bd6b-90351c480fc1","type":"PlayPrompt","branches":[{"condition":"Success","transition":"9b30963a-505a-425a-8a5a-c9fc2e4abd7e"}],"parameters":[{"name":"Text","value":"No option was specified, recording behavior will not be changed.","namespace":null},{"name":"TextToSpeechType","value":"text"}],"metadata":{"position":{"x":919,"y":667},"useDynamic":false}},
        {"id":"443b99d4-3008-4e3b-a51a-8300bcf0817d","type":"PlayPrompt","branches":[{"condition":"Success","transition":"356d5413-950c-43f1-ab5b-bf8b39ee87f4"}],"parameters":[{"name":"Text","value":"This flow will allow you to adjust call recording behavior once this contact is connected to an agent. Note:any recordings will be stored in Amazon S3.","namespace":null},{"name":"TextToSpeechType","value":"text"}],"metadata":{"position":{"x":267,"y":308},"useDynamic":false}},
        {"id":"9b30963a-505a-425a-8a5a-c9fc2e4abd7e","type":"Transfer","branches":[{"condition":"Error","transition":"7cbaf36b-2899-43f5-b834-333a68d2067a"}],"parameters":[{"name":"ContactFlowId","value":"arn:aws:connect:eu-west-2:456789012345:instance/ffffffff-ffff-4000-ffff-ffffffffffff/contact-flow/f047d1a0-c42c-4660-9e9f-fa684b7d4493","resourceName":"Sample inbound flow (first contact experience)"}],"metadata":{"position":{"x":1275,"y":363},"useDynamic":false,"ContactFlow":{"id":"arn:aws:connect:eu-west-2:456789012345:instance/ffffffff-ffff-4000-ffff-ffffffffffff/contact-flow/f047d1a0-c42c-4660-9e9f-fa684b7d4493","text":"Sample inbound flow (first contact experience)"}},"target":"Flow"},
        {"id":"356d5413-950c-43f1-ab5b-bf8b39ee87f4","type":"GetUserInput","branches":[{"condition":"Evaluate","conditionType":"Equals","conditionValue":"1","transition":"8a2cf897-4db1-4524-8acf-d8349bf3b5ee"},{"condition":"Evaluate","conditionType":"Equals","conditionValue":"2","transition":"53aab24b-bbe5-4014-b189-81f7a66a3997"},{"condition":"Evaluate","conditionType":"Equals","conditionValue":"3","transition":"c19afb78-cec9-48e5-b696-22d64f2832f1"},{"condition":"Evaluate","conditionType":"Equals","conditionValue":"4","transition":"e9b9dea3-74d8-4cf5-ac46-fe11369e7037"},{"condition":"Timeout","transition":"5a3c5b08-e33b-4485-bd6b-90351c480fc1"},{"condition":"NoMatch","transition":"5a3c5b08-e33b-4485-bd6b-90351c480fc1"},{"condition":"Error","transition":"5a3c5b08-e33b-4485-bd6b-90351c480fc1"}],"parameters":[{"name":"Text","value":"Press 1 to turn on agent and customer recording. \nPress 2 to turn on agent only recording. \nPress 3 to turn on customer only recording. \nPress 4 to turn off all recording.","namespace":null},{"name":"TextToSpeechType","value":"text"},{"name":"Timeout","value":"8"},{"name":"MaxDigits","value":"1"}],"metadata":{"position":{"x":577,"y":196},"conditionMetadata":[{"id":"7d5e25c8-9610-4101-a65b-58b9f89c0eed","value":"1"},{"id":"9e663d7b-356d-415e-b86b-1331bd2e41a5","value":"2"},{"id":"c40aab24-594f-4f87-8bf6-6681474d38fe","value":"3"},{"id":"36b4e3c9-d48d-4cc2-a2b7-0232be292953","value":"4"}],"useDynamic":false},"target":"Digits"},
        {"id":"8a2cf897-4db1-4524-8acf-d8349bf3b5ee","type":"SetRecordingBehavior","branches":[{"condition":"Success","transition":"9b30963a-505a-425a-8a5a-c9fc2e4abd7e"}],"parameters":[{"name":"RecordingBehaviorOption","value":"Enable"},{"name":"RecordingParticipantOption","value":"Both"}],"metadata":{"position":{"x":917,"y":35}}},
        {"id":"5808b567-aa75-492d-968d-31c48df1b3fe","type":"CheckAttribute","branches":[{"condition":"Evaluate","conditionType":"Equals","conditionValue":"CHAT","transition":"fc597c74-f076-4597-9bf4-f85223746fa3"},{"condition":"NoMatch","transition":"443b99d4-3008-4e3b-a51a-8300bcf0817d"}],"parameters":[{"name":"Attribute","value":"Channel"},{"name":"Namespace","value":"System"}],"metadata":{"position":{"x":30,"y":216},"conditionMetadata":[{"id":"bb2f13f6-2856-4718-b0e9-ce8bc6e58307","operator":{"name":"Equals","value":"Equals","shortDisplay":"="},"value":"CHAT"}]}},
        {"id":"fc597c74-f076-4597-9bf4-f85223746fa3","type":"PlayPrompt","branches":[{"condition":"Success","transition":"8a2cf897-4db1-4524-8acf-d8349bf3b5ee"}],"parameters":[{"name":"Text","value":"For chat, this flow will enable managers to monitor ongoing chat conversations.","namespace":null},{"name":"TextToSpeechType","value":"text"}],"metadata":{"position":{"x":320,"y":39},"useDynamic":false}}
    ],
    "version":"1","type":"contactFlow","start":"5808b567-aa75-492d-968d-31c48df1b3fe",
    "metadata":{"entryPointPosition":{"x":11,"y":24},"snapToGrid":false,"name":"Sample recording behavior","description":"Sample flow to enable recording behavior","type":"contactFlow","status":"published","hash":"db7614bf498026c56280bff5b39ac97c9e3fa073791ae22f6c2b068126412b70"}
}`

var sampleInput = `{
    "modules":[
        {"id":"50b6aab8-40d8-4751-8a14-8f3df571d145","type":"TransferToFlow","branches":[{"condition":"Error","transition":"7670fbff-a8d2-451b-ae8f-546c90e92554"}],"parameters":[{"name":"ContactFlowId","value":"arn:aws:connect:eu-west-2:456789012345:instance/ffffffff-ffff-4000-ffff-ffffffffffff/contact-flow/f047d1a0-c42c-4660-9e9f-fa684b7d4493"}],"metadata":{"position":{"x":1463,"y":297},"useDynamic":false,"ContactFlow":{"id":"arn:aws:connect:eu-west-2:456789012345:instance/ffffffff-ffff-4000-ffff-ffffffffffff/contact-flow/f047d1a0-c42c-4660-9e9f-fa684b7d4493","text":"Sample inbound flow (first contact experience)","arn":null,"metricDetail":null}}},
        {"id":"7670fbff-a8d2-451b-ae8f-546c90e92554","type":"PlayAudio","branches":[{"condition":"Success","transition":"401a0775-88b3-457f-b9e6-f92f13348853"}],"parameters":[{"name":"Text","value":"We are unable to return back to the flow. Goodbye.","namespace":null},{"name":"TextToSpeechType","value":"text"}],"metadata":{"position":{"x":1694,"y":297},"useDynamic":false}},
        {"id":"401a0775-88b3-457f-b9e6-f92f13348853","type":"Disconnect","branches":[],"parameters":[],"metadata":{"position":{"x":1947,"y":387}}},
        {"id":"5f924f08-6a62-4850-80a3-c24a00cc49f1","type":"PlayAudio","branches":[{"condition":"Success","transition":"50b6aab8-40d8-4751-8a14-8f3df571d145"}],"parameters":[{"name":"Text","value":"Returning back to the original flow.","namespace":null},{"name":"TextToSpeechType","value":"text"}],"metadata":{"position":{"x":1228,"y":298},"useDynamic":false}},
        {"id":"3e05fb47-2100-4410-a333-26e12f30429d","type":"PlayAudio","branches":[{"condition":"Success","transition":"5f924f08-6a62-4850-80a3-c24a00cc49f1"}],"parameters":[{"name":"Text","value":"The encrypted customer credit card number is now saved and can be passed to the agent as a screenpop or processed using AWS Lambda. You may also want to check out the sample secure input flow with an agent.","namespace":null},{"name":"TextToSpeechType","value":"text"}],"metadata":{"position":{"x":959,"y":158},"useDynamic":false}},
        {"id":"b092c9fa-ec2f-48c7-a22a-d46184d4af61","type":"SetScreenPop","branches":[{"condition":"Success","transition":"3e05fb47-2100-4410-a333-26e12f30429d"},{"condition":"Error","transition":"e1d99e8e-e457-491b-af02-53d70d90580f"}],"parameters":[{"name":"key","value":"EncryptedCreditCard"},{"name":"value","value":"Stored customer input","namespace":"System"}],"metadata":{"position":{"x":686,"y":156},"useDynamic":true}},
        {"id":"e1d99e8e-e457-491b-af02-53d70d90580f","type":"PlayAudio","branches":[{"condition":"Success","transition":"5f924f08-6a62-4850-80a3-c24a00cc49f1"}],"parameters":[{"name":"Text","value":"There was a problem gathering the customer's input. Did you specify an encryption key in the Store customer input block?","namespace":null},{"name":"TextToSpeechType","value":"text"}],"metadata":{"position":{"x":961,"y":327},"useDynamic":false}},
        {"id":"4b7ebbdb-22ed-4cfd-b566-8285daa53cbc","type":"StoreCustomerInput","branches":[{"condition":"Success","transition":"b092c9fa-ec2f-48c7-a22a-d46184d4af61"},{"condition":"Error","transition":"e1d99e8e-e457-491b-af02-53d70d90580f"}],"parameters":[{"name":"Text","value":"Please enter your credit card number, press the pound key when complete.","namespace":null},{"name":"TextToSpeechType","value":"text"},{"name":"CustomerInputType","value":"Custom"},{"name":"Timeout","value":"6"},{"name":"MaxDigits","value":20},{"name":"EncryptEntry","value":true},{"name":"EncryptionKeyId","value":"your-key-id","namespace":null},{"name":"EncryptionKey","value":"Certificate to use for encryption should be provided here. You will need to also upload a signing key in the AWS console","namespace":null}],"metadata":{"position":{"x":426,"y":242},"useDynamic":false,"useDynamicForEncryptionKeys":false,"countryCodePrefix":"+1"}},
        {"id":"ead55375-e379-4af5-80bc-527d04131fcc","type":"PlayAudio","branches":[{"condition":"Success","transition":"4b7ebbdb-22ed-4cfd-b566-8285daa53cbc"}],"parameters":[{"name":"Text","value":"This flow enables users to enter information secured by an encryption key you provide.","namespace":null},{"name":"TextToSpeechType","value":"text"}],"metadata":{"position":{"x":194,"y":245},"useDynamic":false}}
    ],
    "version":"1", "start":"ead55375-e379-4af5-80bc-527d04131fcc",
    "metadata":{"entryPointPosition":{"x":39,"y":219},"name":"Sample secure input with no agent","description":"Enables the customer to enter digits in private. In a real world implementation, enabling encryption is likely preferred.","type":"contactFlow","status":"published","hash":"cb9659fab04f5bbc30361b8a9f9aa73cea5be1d7e589afa449bfb7bab92008ea"}
}`

var sampleQueue = `{
    "modules":[
        {"id":"2c27f89c-81c0-4110-98c3-db2e2fa15ab7","type":"Disconnect","branches":[],"parameters":[],"metadata":{"position":{"x":1349,"y":275}}},
        {"id":"0ada9849-cd1d-485b-bce4-6e620317c4b1","type":"PlayPrompt","branches":[{"condition":"Success","transition":"2c27f89c-81c0-4110-98c3-db2e2fa15ab7"}],"parameters":[{"name":"Text","value":"We are not able to take your call right now. Goodbye.","namespace":null},{"name":"TextToSpeechType","value":"text"}],"metadata":{"position":{"x":1127,"y":255},"useDynamic":false}},
        {"id":"2c1bb3e4-d5ba-401e-b698-d6a26573c7b3","type":"SetQueue","branches":[{"condition":"Success","transition":"2bfec059-0c2c-45c2-bc42-dfed3501ca2e"},{"condition":"Error","transition":"0ada9849-cd1d-485b-bce4-6e620317c4b1"}],"parameters":[{"name":"Queue","value":"arn:aws:connect:eu-west-2:456789012345:instance/ffffffff-ffff-4000-ffff-ffffffffffff/queue/ff305403-5502-4475-b458-d4bacee82020","namespace":null,"resourceName":"Level 1 Queue"}],"metadata":{"position":{"x":153,"y":39},"useDynamic":false,"queue":{"id":"arn:aws:connect:eu-west-2:456789012345:instance/ffffffff-ffff-4000-ffff-ffffffffffff/queue/ff305403-5502-4475-b458-d4bacee82020","text":"Level 1 Queue"}}},
        {"id":"47480c3a-fb85-45d3-bcbd-42ded95b3724","type":"Transfer","branches":[{"condition":"AtCapacity","transition":"0ada9849-cd1d-485b-bce4-6e620317c4b1"},{"condition":"Error","transition":"0ada9849-cd1d-485b-bce4-6e620317c4b1"}],"parameters":[],"metadata":{"position":{"x":882,"y":34},"useDynamic":false,"queue":null},"target":"Queue"},
        {"id":"2bfec059-0c2c-45c2-bc42-dfed3501ca2e","type":"CheckHoursOfOperation","branches":[{"condition":"True","transition":"b08e042d-af3b-4c8a-bdc7-428c24ffcaf0"},{"condition":"False","transition":"0ada9849-cd1d-485b-bce4-6e620317c4b1"},{"condition":"Error","transition":"0ada9849-cd1d-485b-bce4-6e620317c4b1"}],"parameters":[],"metadata":{"position":{"x":389,"y":62}}},
        {"id":"b08e042d-af3b-4c8a-bdc7-428c24ffcaf0","type":"PlayPrompt","branches":[{"condition":"Success","transition":"47480c3a-fb85-45d3-bcbd-42ded95b3724"}],"parameters":[{"name":"Text","value":"Transferring you to queue.","namespace":null},{"name":"TextToSpeechType","value":"text"}],"metadata":{"position":{"x":624,"y":60},"useDynamic":false}}
    ],
    "version":"1","type":"contactFlow","start":"2c1bb3e4-d5ba-401e-b698-d6a26573c7b3",
    "metadata":{"entryPointPosition":{"x":15,"y":15},"snapToGrid":false,"name":"Sample queue customer","description":"Places the customer in a queue.","type":"contactFlow","status":"published","hash":"dbc4ca98cdd2977247b7d9809073a286c786763befb4490da7b595a26cafbe76"}
}`

var sampleNote = `{
    "modules":[
        {"id":"35dde84a-f902-4a2b-9e0d-b6079d2a5260","type":"CustomerInQueue","branches":[{"condition":"AtCapacity","transition":"e8f664fd-3e86-4aaa-915f-c1fa4a4d4dd0"},{"condition":"Error","transition":"e8f664fd-3e86-4aaa-915f-c1fa4a4d4dd0"}],"parameters":[],"metadata":{"position":{"x":928,"y":85},"useDynamic":false,"queue":null}},
        {"id":"8f9f6aa6-88bc-4bbe-8920-5d2ca2d5cfb4","type":"Disconnect","branches":[],"parameters":[],"metadata":{"position":{"x":1373,"y":231}}},
        {"id":"8595ce23-0af7-49e3-8324-22accebefd0d","type":"SetQueue","branches":[{"condition":"Success","transition":"35dde84a-f902-4a2b-9e0d-b6079d2a5260"},{"condition":"Error","transition":"e8f664fd-3e86-4aaa-915f-c1fa4a4d4dd0"}],"parameters":[{"name":"Queue","value":"arn:aws:connect:eu-west-2:456789012345:instance/ffffffff-ffff-4000-ffff-ffffffffffff/queue/ff305403-5502-4475-b458-d4bacee82020","namespace":null}],"metadata":{"position":{"x":696,"y":72},"useDynamic":false,"queue":{"id":"arn:aws:connect:eu-west-2:456789012345:instance/ffffffff-ffff-4000-ffff-ffffffffffff/queue/ff305403-5502-4475-b458-d4bacee82020","text":"BasicQueue","arn":null,"metricDetail":null}}},
        {"id":"e8f664fd-3e86-4aaa-915f-c1fa4a4d4dd0","type":"PlayAudio","branches":[{"condition":"Success","transition":"8f9f6aa6-88bc-4bbe-8920-5d2ca2d5cfb4"}],"parameters":[{"name":"Text","value":"An error ocurred. Goodbye.","namespace":null},{"name":"TextToSpeechType","value":"text"}],"metadata":{"position":{"x":1147,"y":155},"useDynamic":false}},
        {"id":"e1aff3c4-01f1-43aa-864e-371a6af4095f","type":"SetScreenPop","branches":[{"condition":"Success","transition":"8595ce23-0af7-49e3-8324-22accebefd0d"},{"condition":"Error","transition":"e8f664fd-3e86-4aaa-915f-c1fa4a4d4dd0"}],"parameters":[{"name":"key","value":"note"},{"name":"value","value":"This note demonstrates how attributes can appear in the screen pop.","namespace":null}],"metadata":{"position":{"x":448,"y":94},"useDynamic":false}},
        {"id":"2b6ee9f2-0f1a-417c-91c7-bdc12cfc6a07","type":"PlayAudio","branches":[{"condition":"Success","transition":"e1aff3c4-01f1-43aa-864e-371a6af4095f"}],"parameters":[{"name":"Text","value":"This sets a note attribute for use in a screenpop.","namespace":null},{"name":"TextToSpeechType","value":"text"}],"metadata":{"position":{"x":209,"y":61},"useDynamic":false}}
    ],
    "version":"1","start":"2b6ee9f2-0f1a-417c-91c7-bdc12cfc6a07",
    "metadata":{"entryPointPosition":{"x":75,"y":20},"name":"Sample note for screenpop","description":"Screenpop is a Contact control pannel feature that allows loading a web page optionally with parameters based on attributes. Refer to the screenpop documentation for more information.","type":"contactFlow","status":"published","hash":"58fe14cc6789928b2b0387fbc86528acd84e455237d72e54e30af188211a20a7"}
}`

var sampleAB = `{
    "modules":[
        {"id":"1300053c-a75c-462a-9ece-46b0f5c8084a","type":"PlayAudio","branches":[{"condition":"Success","transition":"80329f54-d64f-4510-8eee-a935e4ae0e2c"}],"parameters":[{"name":"Text","value":"Amazon Connect will now simulate rolling dice by using the Distribute randomly block,,,now rolling,,,,,,,","namespace":null},{"name":"TextToSpeechType","value":"text"}],"metadata":{"position":{"x":150,"y":167},"useDynamic":false}},
        {"id":"80329f54-d64f-4510-8eee-a935e4ae0e2c","type":"PercentileBranching","branches":[{"condition":"Evaluate","conditionType":"Equals","conditionValue":"3","transition":"034ceb17-23e9-4675-a437-41a89c1568e8"},{"condition":"Evaluate","conditionType":"Equals","conditionValue":"6","transition":"c9c6df6c-1fb4-459b-b84d-28f92d0c1422"},{"condition":"Evaluate","conditionType":"Equals","conditionValue":"8","transition":"6676b92a-7e0b-47a5-b55c-56d31e465218"},{"condition":"Evaluate","conditionType":"Equals","conditionValue":"11","transition":"1187f441-b61e-420e-a554-6799893e3e50"},{"condition":"Evaluate","conditionType":"Equals","conditionValue":"14","transition":"aaffb149-bd95-4a2f-92ed-12bab147442f"},{"condition":"Evaluate","conditionType":"Equals","conditionValue":"17","transition":"42520dfc-a90d-40b3-a1c6-92bc62250e24"},{"condition":"Evaluate","conditionType":"Equals","conditionValue":"14","transition":"89660d4b-56fe-4ec5-8038-97c910dff3be"},{"condition":"Evaluate","conditionType":"Equals","conditionValue":"11","transition":"1892aca2-3226-4249-aafa-fe68b8b1c2de"},{"condition":"Evaluate","conditionType":"Equals","conditionValue":"8","transition":"cc35093a-2118-4d5e-a689-2c26be2aee76"},{"condition":"Evaluate","conditionType":"Equals","conditionValue":"5","transition":"b20bf80e-ea4c-4914-b4e5-f0865c0aaab8"},{"conditionType":"Equals","conditionValue":"3","condition":"NoMatch","transition":"5924eb92-9d3b-4d0c-b5a6-76b589779c97"}],"parameters":[],"metadata":{"position":{"x":383,"y":337},"conditionMetadata":[{"id":"6a5c8eab-c3b0-4515-afa8-9a1a307e133d","percent":{"value":1,"display":"1%"},"name":"2","value":"3"},{"id":"f0c3a494-4d54-4d35-be5f-183016dc2a98","percent":{"value":1,"display":"1%"},"name":"3","value":"6"},{"id":"2789b854-87f8-48cf-84a4-708d35c1d52a","percent":{"value":1,"display":"1%"},"name":"4","value":"8"},{"id":"03c2abed-62b8-4a61-965b-112a8a49b3b2","percent":{"value":1,"display":"1%"},"name":"5","value":"11"},{"id":"680f87d5-7afe-4402-b8c1-38f30a1f4bc8","percent":{"value":1,"display":"1%"},"name":"6","value":"14"},{"id":"30beeb59-b040-4993-a3e3-2564223913dc","percent":{"value":1,"display":"1%"},"name":"7","value":"17"},{"id":"73dd2d52-0d57-47d8-b515-17da5f9fb6e6","percent":{"value":1,"display":"1%"},"name":"8","value":"14"},{"id":"b384a3e7-5967-4fa4-9eef-29889b8a3c42","percent":{"value":1,"display":"1%"},"name":"9","value":"11"},{"id":"2179c9b9-321e-4ba1-8e22-00213397923f","percent":{"value":1,"display":"1%"},"name":"10","value":"8"},{"id":"b71a6b53-aa3e-4898-aaa8-79d51645a37c","percent":{"value":1,"display":"1%"},"name":"11","value":"5"}]}},
        {"id":"bdc8e490-d2e1-4e45-9c5f-34612f60a846","type":"Disconnect","branches":[],"parameters":[],"metadata":{"position":{"x":1627,"y":807}}},
        {"id":"be80759a-bfdb-4332-bbac-63ddd9ad4dea","type":"TransferToFlow","branches":[{"condition":"Error","transition":"bdc8e490-d2e1-4e45-9c5f-34612f60a846"}],"parameters":[{"name":"ContactFlowId","value":"arn:aws:connect:eu-west-2:456789012345:instance/ffffffff-ffff-4000-ffff-ffffffffffff/contact-flow/f047d1a0-c42c-4660-9e9f-fa684b7d4493"}],"metadata":{"position":{"x":1392,"y":802},"useDynamic":false,"ContactFlow":{"id":"arn:aws:connect:eu-west-2:456789012345:instance/ffffffff-ffff-4000-ffff-ffffffffffff/contact-flow/f047d1a0-c42c-4660-9e9f-fa684b7d4493","text":"Sample inbound flow (first contact experience)","arn":null,"metricDetail":null}}},
        {"id":"034ceb17-23e9-4675-a437-41a89c1568e8","type":"PlayAudio","branches":[{"condition":"Success","transition":"f1bd6ea7-c8e9-4d2f-9c1e-13aab35caf77"}],"parameters":[{"name":"Text","value":"You rolled a 2!","namespace":null},{"name":"TextToSpeechType","value":"text"}],"metadata":{"position":{"x":719,"y":12},"useDynamic":false}},
        {"id":"c9c6df6c-1fb4-459b-b84d-28f92d0c1422","type":"PlayAudio","branches":[{"condition":"Success","transition":"f1bd6ea7-c8e9-4d2f-9c1e-13aab35caf77"}],"parameters":[{"name":"Text","value":"You rolled a 3!","namespace":null},{"name":"TextToSpeechType","value":"text"}],"metadata":{"position":{"x":719,"y":162},"useDynamic":false}},
        {"id":"6676b92a-7e0b-47a5-b55c-56d31e465218","type":"PlayAudio","branches":[{"condition":"Success","transition":"f1bd6ea7-c8e9-4d2f-9c1e-13aab35caf77"}],"parameters":[{"name":"Text","value":"You rolled a 4!","namespace":null},{"name":"TextToSpeechType","value":"text"}],"metadata":{"position":{"x":720,"y":308},"useDynamic":false}},
        {"id":"1187f441-b61e-420e-a554-6799893e3e50","type":"PlayAudio","branches":[{"condition":"Success","transition":"f1bd6ea7-c8e9-4d2f-9c1e-13aab35caf77"}],"parameters":[{"name":"Text","value":"You rolled a 5!","namespace":null},{"name":"TextToSpeechType","value":"text"}],"metadata":{"position":{"x":720,"y":452},"useDynamic":false}},
        {"id":"aaffb149-bd95-4a2f-92ed-12bab147442f","type":"PlayAudio","branches":[{"condition":"Success","transition":"f1bd6ea7-c8e9-4d2f-9c1e-13aab35caf77"}],"parameters":[{"name":"Text","value":"You rolled a 6!","namespace":null},{"name":"TextToSpeechType","value":"text"}],"metadata":{"position":{"x":721,"y":593},"useDynamic":false}},
        {"id":"42520dfc-a90d-40b3-a1c6-92bc62250e24","type":"PlayAudio","branches":[{"condition":"Success","transition":"f1bd6ea7-c8e9-4d2f-9c1e-13aab35caf77"}],"parameters":[{"name":"Text","value":"You rolled a 7!","namespace":null},{"name":"TextToSpeechType","value":"text"}],"metadata":{"position":{"x":722,"y":735},"useDynamic":false}},
        {"id":"89660d4b-56fe-4ec5-8038-97c910dff3be","type":"PlayAudio","branches":[{"condition":"Success","transition":"f1bd6ea7-c8e9-4d2f-9c1e-13aab35caf77"}],"parameters":[{"name":"Text","value":"You rolled a 8!","namespace":null},{"name":"TextToSpeechType","value":"text"}],"metadata":{"position":{"x":723,"y":879},"useDynamic":false}},
        {"id":"1892aca2-3226-4249-aafa-fe68b8b1c2de","type":"PlayAudio","branches":[{"condition":"Success","transition":"f1bd6ea7-c8e9-4d2f-9c1e-13aab35caf77"}],"parameters":[{"name":"Text","value":"You rolled a 9!","namespace":null},{"name":"TextToSpeechType","value":"text"}],"metadata":{"position":{"x":723,"y":1021},"useDynamic":false}},
        {"id":"cc35093a-2118-4d5e-a689-2c26be2aee76","type":"PlayAudio","branches":[{"condition":"Success","transition":"f1bd6ea7-c8e9-4d2f-9c1e-13aab35caf77"}],"parameters":[{"name":"Text","value":"You rolled a 10!","namespace":null},{"name":"TextToSpeechType","value":"text"}],"metadata":{"position":{"x":724,"y":1163},"useDynamic":false}},
        {"id":"b20bf80e-ea4c-4914-b4e5-f0865c0aaab8","type":"PlayAudio","branches":[{"condition":"Success","transition":"f1bd6ea7-c8e9-4d2f-9c1e-13aab35caf77"}],"parameters":[{"name":"Text","value":"You rolled a 11!","namespace":null},{"name":"TextToSpeechType","value":"text"}],"metadata":{"position":{"x":725,"y":1305},"useDynamic":false}},
        {"id":"5924eb92-9d3b-4d0c-b5a6-76b589779c97","type":"PlayAudio","branches":[{"condition":"Success","transition":"f1bd6ea7-c8e9-4d2f-9c1e-13aab35caf77"}],"parameters":[{"name":"Text","value":"You rolled a 12!","namespace":null},{"name":"TextToSpeechType","value":"text"}],"metadata":{"position":{"x":725,"y":1448},"useDynamic":false}},
        {"id":"f1bd6ea7-c8e9-4d2f-9c1e-13aab35caf77","type":"PlayAudio","branches":[{"condition":"Success","transition":"be80759a-bfdb-4332-bbac-63ddd9ad4dea"}],"parameters":[{"name":"Text","value":"Now transferring back to the main menu.","namespace":null},{"name":"TextToSpeechType","value":"text"}],"metadata":{"position":{"x":1128,"y":1518},"useDynamic":false} }
    ],
    "version":"1","start":"1300053c-a75c-462a-9ece-46b0f5c8084a",
    "metadata":{"entryPointPosition":{"x":17,"y":126},"name":"Sample AB test","description":"Performs A/B call distribution","type":"contactFlow","status":"published","hash":"ec3bfccf5ff85706d61890b102758b73e76bf63c7690025b235182717e13d935"}
}`

var sampleQueueConfig = `{
    "modules":[
        {"id":"15445378-0358-427c-8e3e-d4693a07cf8d","type":"PlayPrompt","branches":[{"condition":"Success","transition":"05c9b198-6508-4fb2-af45-8526c535a05c"}],"parameters":[{"name":"Text","value":"The time in queue is more than 5 minutes.","namespace":null},{"name":"TextToSpeechType","value":"text"}],"metadata":{"position":{"x":1077,"y":178},"useDynamic":false}},
        {"id":"dbb742ec-5005-403b-896b-30b7cbf63732","type":"SetContactFlow","branches":[{"condition":"Success","transition":"1761a5fd-0eab-46c6-801f-1df5ec18ea56"}],"parameters":[{"name":"ContactFlowId","value":"arn:aws:connect:eu-west-2:188451104469:instance/7b5e19d1-2e3a-4c77-9cab-2aed70656981/contact-flow/0e9e2db9-411e-4329-8380-f878fdefcbf0","resourceName":"Default customer queue"},{"name":"Type","value":"CustomerQueue"}],"metadata":{"position":{"x":41,"y":375},"contactFlow":{"id":"arn:aws:connect:eu-west-2:188451104469:instance/7b5e19d1-2e3a-4c77-9cab-2aed70656981/contact-flow/0e9e2db9-411e-4329-8380-f878fdefcbf0","text":"Default customer queue"},"customerOrAgent":false}},
        {"id":"67f2d6ed-c5bc-42cf-9011-b0589ef14103","type":"PlayPrompt","branches":[{"condition":"Success","transition":"c69c6e6e-8707-42e5-b8bc-d79d64385234"}],"parameters":[{"name":"Text","value":"You are now being placed in queue to chat with an agent.","namespace":null},{"name":"TextToSpeechType","value":"text"}],"metadata":{"position":{"x":1664,"y":65},"useDynamic":false}},
        {"id":"da46c982-772f-43be-a034-8402e4d7afea","type":"Transfer","branches":[{"condition":"AtCapacity","transition":"43d31362-47a0-4897-8592-32ffc70b3b3c"},{"condition":"Error","transition":"43d31362-47a0-4897-8592-32ffc70b3b3c"}],"parameters":[],"metadata":{"position":{"x":1646,"y":564},"useDynamic":false,"queue":null},"target":"Queue"},
        {"id":"c69c6e6e-8707-42e5-b8bc-d79d64385234","type":"Transfer","branches":[{"condition":"AtCapacity","transition":"43d31362-47a0-4897-8592-32ffc70b3b3c"},{"condition":"Error","transition":"43d31362-47a0-4897-8592-32ffc70b3b3c"}],"parameters":[],"metadata":{"position":{"x":1926,"y":39},"useDynamic":false,"queue":null},"target":"Queue"},
        {"id":"84e7e723-838f-4012-bafe-ffbda3b81d51","type":"PlayPrompt","branches":[{"condition":"Success","transition":"05c9b198-6508-4fb2-af45-8526c535a05c"}],"parameters":[{"name":"Text","value":"The time in queue is less than 5 minutes.","namespace":null},{"name":"TextToSpeechType","value":"text"}],"metadata":{"position":{"x":1079,"y":14},"useDynamic":false}},
        {"id":"049d9571-48bf-4fa9-9b63-683288ddb456","type":"SetContactFlow","branches":[{"condition":"Success","transition":"da46c982-772f-43be-a034-8402e4d7afea"}],"parameters":[{"name":"ContactFlowId","value":"arn:aws:connect:eu-west-2:188451104469:instance/7b5e19d1-2e3a-4c77-9cab-2aed70656981/contact-flow/2513074e-653e-4d32-a5b0-ad3fa15bf760","resourceName":"Sample interruptible queue flow with callback"},{"name":"Type","value":"CustomerQueue"}],"metadata":{"position":{"x":1412,"y":517},"contactFlow":{"id":"arn:aws:connect:eu-west-2:188451104469:instance/7b5e19d1-2e3a-4c77-9cab-2aed70656981/contact-flow/2513074e-653e-4d32-a5b0-ad3fa15bf760","text":"Sample interruptible queue flow with callback"},"customerOrAgent":false}},
        {"id":"05c9b198-6508-4fb2-af45-8526c535a05c","type":"CheckAttribute","branches":[{"condition":"Evaluate","conditionType":"Equals","conditionValue":"CHAT","transition":"67f2d6ed-c5bc-42cf-9011-b0589ef14103"},{"condition":"NoMatch","transition":"35d19e2a-432e-41a6-838d-bb9949910869"}],"parameters":[{"name":"Attribute","value":"Channel"},{"name":"Namespace","value":"System"}],"metadata":{"position":{"x":1337,"y":37},"conditionMetadata":[{"id":"8476ed95-f3f6-4c70-90ea-a195759ecd90","operator":{"name":"Equals","value":"Equals","shortDisplay":"="},"value":"CHAT"}]}},
        {"id":"f48561e5-3d02-4380-8031-8139cee399d2","type":"Disconnect","branches":[],"parameters":[],"metadata":{"position":{"x":890,"y":709}}},
        {"id":"5cfb3fa1-78a7-4133-b553-ca7198181192","type":"SetQueue","branches":[{"condition":"Success","transition":"dbb742ec-5005-403b-896b-30b7cbf63732"},{"condition":"Error","transition":"cad6d3f7-e39e-426f-bf32-b99aa549a34b"}],"parameters":[{"name":"Queue","value":"arn:aws:connect:eu-west-2:188451104469:instance/7b5e19d1-2e3a-4c77-9cab-2aed70656981/queue/ff305403-5502-4475-b458-d4bacee82020","namespace":null,"resourceName":"BasicQueue"}],"metadata":{"position":{"x":31,"y":158},"useDynamic":false,"queue":{"id":"arn:aws:connect:eu-west-2:188451104469:instance/7b5e19d1-2e3a-4c77-9cab-2aed70656981/queue/ff305403-5502-4475-b458-d4bacee82020","text":"BasicQueue"}}},
        {"id":"1761a5fd-0eab-46c6-801f-1df5ec18ea56","type":"CheckHoursOfOperation","branches":[{"condition":"True","transition":"3746a42c-0735-4668-ade2-9335496f495e"},{"condition":"False","transition":"cad6d3f7-e39e-426f-bf32-b99aa549a34b"},{"condition":"Error","transition":"cad6d3f7-e39e-426f-bf32-b99aa549a34b"}],"parameters":[],"metadata":{"position":{"x":48,"y":558}}},
        {"id":"a91b0f5f-71dc-4d99-adbe-827db4b13591","type":"UpdateRoutingPriority","branches":[{"condition":"Success","transition":"3c52961e-581a-42b5-87b9-41c7fee5dbf9"}],"parameters":[{"name":"AbsolutePosition","value":1}],"metadata":{"position":{"x":592,"y":239},"adjustUnit":null}},
        {"id":"76f56182-8786-42d7-960e-d99b54a45a45","type":"UpdateRoutingPriority","branches":[{"condition":"Success","transition":"3c52961e-581a-42b5-87b9-41c7fee5dbf9"}],"parameters":[{"name":"TimeOffset","value":600}],"metadata":{"position":{"x":589,"y":404},"adjustUnit":"minutes"}},
        {"id":"3c52961e-581a-42b5-87b9-41c7fee5dbf9","type":"CheckQueueStatus","branches":[{"condition":"CheckQueueTime","conditionType":"LessThan","conditionValue":300000,"transition":"84e7e723-838f-4012-bafe-ffbda3b81d51"},{"condition":"NoMatch","transition":"15445378-0358-427c-8e3e-d4693a07cf8d"},{"condition":"Error","transition":"cad6d3f7-e39e-426f-bf32-b99aa549a34b"}],"parameters":[],"metadata":{"position":{"x":818,"y":25},"conditionMetadata":[{"id":"289c7f2e-a8ac-4d99-9525-3a3064094073","attribute":{"name":"Time in Queue","value":"Time in Queue"},"operator":{"name":"Is less than","value":"LessThan","shortDisplay":"<"},"value":"300","time":"sec."}],"queue":null,"useDynamic":false}},
        {"id":"cad6d3f7-e39e-426f-bf32-b99aa549a34b","type":"PlayPrompt","branches":[{"condition":"Success","transition":"f48561e5-3d02-4380-8031-8139cee399d2"}],"parameters":[{"name":"Text","value":"We are not able to take your call right now. Goodbye.","namespace":null},{"name":"TextToSpeechType","value":"text"}],"metadata":{"position":{"x":604,"y":698},"useDynamic":false}},
        {"id":"ec7862fb-30fc-45b3-8ac6-776cbe055068","type":"GetUserInput","branches":[{"condition":"Evaluate","conditionType":"Equals","conditionValue":"1","transition":"a91b0f5f-71dc-4d99-adbe-827db4b13591"},{"condition":"Evaluate","conditionType":"Equals","conditionValue":"2","transition":"76f56182-8786-42d7-960e-d99b54a45a45"},{"condition":"Timeout","transition":"3c52961e-581a-42b5-87b9-41c7fee5dbf9"},{"condition":"NoMatch","transition":"3c52961e-581a-42b5-87b9-41c7fee5dbf9"},{"condition":"Error","transition":"cad6d3f7-e39e-426f-bf32-b99aa549a34b"}],"parameters":[{"name":"Text","value":"Press 1 to move to the front of the queue or press 2 to move behind existing contacts already in queue.","namespace":null},{"name":"TextToSpeechType","value":"text"},{"name":"Timeout","value":"5"},{"name":"MaxDigits","value":"1"}],"metadata":{"position":{"x":329,"y":413},"conditionMetadata":[{"id":"33f902ce-9188-436d-a273-ab9557774997","value":"1"},{"id":"f82a88fc-b93a-4e17-93e4-b016b1cc3d03","value":"2"}],"useDynamic":false},"target":"Digits"},
        {"id":"3746a42c-0735-4668-ade2-9335496f495e","type":"CheckAttribute","branches":[{"condition":"Evaluate","conditionType":"Equals","conditionValue":"CHAT","transition":"3c52961e-581a-42b5-87b9-41c7fee5dbf9"},{"condition":"NoMatch","transition":"80341111-9853-40db-aa90-2f9dcff12344"}],"parameters":[{"name":"Attribute","value":"Channel"},{"name":"Namespace","value":"System"}],"metadata":{"position":{"x":249,"y":7},"conditionMetadata":[{"id":"96033dbe-420d-4962-b8ad-baf9852bd6a3","operator":{"name":"Equals","value":"Equals","shortDisplay":"="},"value":"CHAT"}]}},
        {"id":"80341111-9853-40db-aa90-2f9dcff12344","type":"PlayPrompt","branches":[{"condition":"Success","transition":"ec7862fb-30fc-45b3-8ac6-776cbe055068"}],"parameters":[{"name":"Text","value":"This flow demonstrates changing the priority of an individual contact in the queue and will allow you to request a callback and be called when an agent is available.","namespace":null},{"name":"TextToSpeechType","value":"text"}],"metadata":{"position":{"x":297,"y":237},"useDynamic":false}},
        {"id":"35d19e2a-432e-41a6-838d-bb9949910869","type":"GetUserInput","branches":[{"condition":"Evaluate","conditionType":"Equals","conditionValue":"2","transition":"220a5a3b-1792-40dd-9c08-08b6bc122541"},{"condition":"Evaluate","conditionType":"Equals","conditionValue":"1","transition":"049d9571-48bf-4fa9-9b63-683288ddb456"},{"condition":"Timeout","transition":"049d9571-48bf-4fa9-9b63-683288ddb456"},{"condition":"NoMatch","transition":"049d9571-48bf-4fa9-9b63-683288ddb456"},{"condition":"Error","transition":"cad6d3f7-e39e-426f-bf32-b99aa549a34b"}],"parameters":[{"name":"Text","value":"Press 1 to go into queue or 2 to enter a callback number.","namespace":null},{"name":"TextToSpeechType","value":"text"},{"name":"Timeout","value":"5"},{"name":"MaxDigits","value":"1"}],"metadata":{"position":{"x":1119,"y":381},"conditionMetadata":[{"id":"222ef4a0-84dc-44e4-a112-6d9b053e415b","value":"2"},{"id":"924e60eb-ff16-49e1-b916-59b10fbe945e","value":"1"}],"useDynamic":false},"target":"Digits"},
        {"id":"58ecb14e-0970-4d5b-9d4c-f2cb7d23bf12","type":"PlayPrompt","branches":[{"condition":"Success","transition":"220a5a3b-1792-40dd-9c08-08b6bc122541"}],"parameters":[{"name":"Text","value":"The number entered is invalid. Please try again.","namespace":null},{"name":"TextToSpeechType","value":"text"}],"metadata":{"position":{"x":1888,"y":415},"useDynamic":false}},
        {"id":"220a5a3b-1792-40dd-9c08-08b6bc122541","type":"StoreUserInput","branches":[{"condition":"Success","transition":"9155d755-b5fb-4c17-a560-8a9d1095eedf"},{"condition":"InvalidNumber","transition":"58ecb14e-0970-4d5b-9d4c-f2cb7d23bf12"},{"condition":"Error","transition":"43d31362-47a0-4897-8592-32ffc70b3b3c"}],"parameters":[{"name":"Text","value":"Enter the number you would like to be called back at.","namespace":null},{"name":"TextToSpeechType","value":"text"},{"name":"CustomerInputType","value":"PhoneNumber"},{"name":"Timeout","value":"6"},{"name":"PhoneNumberFormat","value":"Local"},{"name":"CountryCode","value":"US"}],"metadata":{"position":{"x":1404,"y":270},"useDynamic":false,"useDynamicForEncryptionKeys":true,"countryCodePrefix":"+1"}},
        {"id":"9155d755-b5fb-4c17-a560-8a9d1095eedf","type":"SetCallBackNumber","branches":[{"condition":"Success","transition":"b5a7030b-23c9-4570-80d2-a9ec4c17e806"},{"condition":"InvalidPhoneNumber","transition":"58ecb14e-0970-4d5b-9d4c-f2cb7d23bf12"},{"condition":"NonDialableNumber","transition":"58ecb14e-0970-4d5b-9d4c-f2cb7d23bf12"}],"parameters":[{"name":"CallBackNumber","value":"Stored customer input","namespace":"System"}],"metadata":{"position":{"x":1645,"y":265}}},
        {"id":"0af94805-014a-4fb1-8a35-070fa35e2e22","type":"Disconnect","branches":[],"parameters":[],"metadata":{"position":{"x":2493,"y":379}}},
        {"id":"43d31362-47a0-4897-8592-32ffc70b3b3c","type":"PlayPrompt","branches":[{"condition":"Success","transition":"0af94805-014a-4fb1-8a35-070fa35e2e22"}],"parameters":[{"name":"Text","value":"Something went wrong. Goodbye.","namespace":null},{"name":"TextToSpeechType","value":"text"}],"metadata":{"position":{"x":2182,"y":382},"useDynamic":false}},
        {"id":"b5a7030b-23c9-4570-80d2-a9ec4c17e806","type":"PlayPrompt","branches":[{"condition":"Success","transition":"9ec9348c-c517-403f-935b-9c3bb8e35c1b"}],"parameters":[{"name":"Text","value":"You will be called back when the next agent is available.","namespace":null},{"name":"TextToSpeechType","value":"text"}],"metadata":{"position":{"x":1882,"y":241},"useDynamic":false}},
        {"id":"9ec9348c-c517-403f-935b-9c3bb8e35c1b","type":"CreateCallback","branches":[{"condition":"Success","transition":"0af94805-014a-4fb1-8a35-070fa35e2e22"},{"condition":"Error","transition":"43d31362-47a0-4897-8592-32ffc70b3b3c"}],"parameters":[{"name":"InitialDelaySeconds","value":5},{"name":"RetryDelaySeconds","value":600},{"name":"MaxRetryAttempts","value":1}],"metadata":{"position":{"x":2206,"y":74},"useDynamic":false,"queue":null}}
    ],
    "version":"1","type":"contactFlow","start":"5cfb3fa1-78a7-4133-b553-ca7198181192",
    "metadata":{"entryPointPosition":{"x":20,"y":20},"snapToGrid":false,"name":"Sample queue configurations flow","description":"Puts a customer in queue and gives them the option to be first in queue, last in queue or to be called back.","type":"contactFlow","status":"published","hash":"d08cf945ba9f6f25b6c7a2a4990c48648a2fac8dd66bccfc73ab9c97337627e2"}
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
	for _, flow := range []string{sampleLambda, sampleRecording, sampleInput, sampleQueue, sampleNote, sampleAB, sampleQueueConfig} {
		err = sim.LoadFlowJSON([]byte(flow))
		if err != nil {
			t.Fatalf("unexpected error parsing flow: %v", err)
		}
	}

	type stateLookupOutput struct {
		State string `json:"State"`
	}

	// Register a lambda.
	err = sim.RegisterLambda("state-lookup", func(ctx context.Context, in LambdaPayload) (out stateLookupOutput, err error) {
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

	var call *Call

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

	// Set a custom encryption function.
	sim.SetEncryption(func(in string) string {
		return fmt.Sprintf("(I am encrypting)>༼ つ ◕_◕ ༽つ%s", in)
	})

	// Start a call.
	call, err = sim.StartCall(CallConfig{
		SourceNumber: "+447878123456",
		DestNumber:   "+441121234567",
	})
	if err != nil {
		t.Fatalf("unexpected error starting call: %v", err)
	}

	// Test flows with testing utility.
	expect := NewTestHelper(t, call)

	expect.Prompt().WithPlaintext().Never().ToContain("Error")
	expect.Lambda().WithARN("self-destruct").Never().ToBeInvoked()

	expect.Prompt().WithPlaintext().ToEqual("Hello, thanks for calling. These are some examples of what the Amazon Connect virtual contact center can enable you to do.")
	expect.Prompt().Not().WithSSML().ToContain("3 to hear the results of an AWS Lambda data dip")
	expect.ToEnter("3")
	expect.Prompt().ToContain("Now performing a data dip using AWS Lambda.")
	expect.Lambda().WithParameters(map[string]string{"butter": "salted"}).Not().ToBeInvoked()
	expect.Lambda().WithARN("state-lookup").Not().WithARN("clearly-not-this-one").ToBeInvoked()
	expect.Prompt().ToEqual("Based on the number you are calling from, your area code is located in United Kingdom")
	expect.Prompt().ToEqual("Now returning you to the main menu.")
	expect.Prompt().ToContain("Press 1 to be put in queue for an agent")
	call.Terminate()

	// Run more tests.
	testCases := []struct {
		desc   string
		conf   CallConfig
		assert func(expect *TestHelper)
	}{
		{
			desc: "data entry",
			conf: CallConfig{SourceNumber: "+447878123456", DestNumber: "+441121234567"},
			assert: func(expect *TestHelper) {
				expect.Prompt().ToContain("2 to securely enter content")
				expect.ToEnter("2")
				expect.Prompt().Not().ToContain("error")
				expect.Prompt().ToEqual("This flow enables users to enter information secured by an encryption key you provide.")
				expect.Prompt().ToContain("Please enter your credit card number")
				expect.ToEnter("1234098712340987#")
				expect.UserAttributeUpdate("EncryptedCreditCard", "(I am encrypting)>༼ つ ◕_◕ ༽つ1234098712340987")
				expect.Transfer().ToFlow("Sample inbound flow (first contact experience)")
			},
		},
		{
			desc: "queue transfer",
			conf: CallConfig{SourceNumber: "+447878123456", DestNumber: "+441121234567"},
			assert: func(expect *TestHelper) {
				expect.Prompt().ToContain("4 to set a screen pop for the agent")
				expect.ToEnter("4")
				expect.Prompt().ToEqual("This sets a note attribute for use in a screenpop.")
				expect.Transfer().ToQueue("BasicQueue")
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			call, err := sim.StartCall(tC.conf)
			if err != nil {
				t.Fatalf("unexpected error starting call: %v", err)
			}
			expect := NewTestHelper(t, call)
			tC.assert(expect)
			call.Terminate()
		})
	}
}
