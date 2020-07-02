# amazon-connect-simulator

The aim of this project an emulator for an Amazon Connect instance that can be dropped into your go projects and run on your local machine. This can be used for manually or automatically testing call flows and their interaction with lambda functions. Where Connect deals in spoken words, the simulator uses text for ease of use.

## Features

The simulator is loaded with any flows exported from Amazon Connect. It can accurately simulate:

* Interact: `Play Prompt`, `Get Customer Input`, `Store Customer Input`
* Set: `Set Working Queue`, `Set Contact Attributes`, `Set Voice`
* Branch: `Check Hours Of Operation`, `Check Contact Attributes`
* Integrate: `Invoke AWS Lambda Function`
* Transfer: `Disconnect`, `Transfer To Queue`, `Transfer To Phone Number`, `Transfer To Flow`

For any blocks not on that list, they will be ignored and the flow will continue down the `Success` branch if the block has one. If an unknown block type does not have a success output, the call will terminate at that block.

The following connect features are _not_ presently supported:
* Lex bots
* Pre-recorded prompts
* Queue, Whisper, Outbound, Hold flows etc.
* Interactions with agents, including quick-connect flows
* Text chats

(Amongst other things)

## Setting up a simulator

### Basic Setup

```go
    import (
        simulator "github.com/edwardbrowncross/amazon-connect-simulator"
    )

    func main () {
        // Create a simulator instance.
        sim := simulator.New()

        // Load flow JSON(s). This probably comes from a file on disk.
        err := sim.LoadFlowJSON(fileContents)
		if err != nil {
			panic(err)
        }
        
        // Register phone number(s).
        err = sim.SetStartingFlowFor("+44113123456", "Sample inbound flow (first contact experience)")
        if err != nil {
            panic(err)
        }

        // Start a call.
        // Many calls can be created and run simultaneously from a single simulator instnace.
        call, err := sim.StartCall(simulator.CallConfig{
            SourceNumber: "+447878987654",
            DestNumber: "+44113123456", // Same as registered number.
        })
        if err != nil {
            panic(err)
        }

        // Interact with that call.
        for {
            // Read spoken output.
            prompt, ok := <-call.Caller.O 
            if !ok {
                fmt.Println("call ended")
                break
            }
            fmt.Println(prompt)
            if prompt == "Press 1 to talk to an agent" {
                // Press keypad digits.
                call.Call.I <- '1'
            }
        }
    }
```

### Using Lambdas

Lambda invocation is simulated by Go functions. For each lambda referenced by your flow, you should specify a function to be run for that lambda. The function signature matches that of [a lambda handler](https://github.com/aws/aws-lambda-go/blob/master/events/README_Connect.md) passed to `lambda.Start`. If you wrote your lambdas in Go, then you should be able to use the real lambda handlers in the simulator.

```go
sim := simulator.New()

// The event input type may be any struct.
// The response data type may be anything that can be marshaled as an object in JSON (map[string]string, struct).
// Context and error are non-optional.
greetHandler := func(ctx context.Context, evt events.ConnectEvent) (events.ConnectResponse, error) {
    return events.ConnectResponse{
        "greeting": "Good morning"
    }
}

// The name is any substring of the ARN of the lambda.
sim.registerLambda("get-greeting", greetHandler)

// Using real handlers from elsewhere in your project may be desirable.
sim.registerLambda("account-number", accountlambda.NewHandler(myMockedDependency))
```

### Advanced configuration

A number of other aspects of connect can be mocked with some configured functions.

```go
sim := simulator.New()

// Input block encryption functionality is not provided with this simulator. You may implement your own.
sim.SetEncryption(func(in string, keyID string, cert []byte) (encrypted []byte) {
    return []byte(fmt.Sprintf(`<encrypted key="%s">%s</encrypted>`), keyID, in)
})

// Adds logic used by the checkHoursOfOperation block to determine if we are in operating hours.
// The first parameter of the provided function will either be the name of a Queue or the name of an Hours of Operation, as indicated by the second parameter.
// Returning an error indicates that the given queue/hours does not exist or does not have hours defined. The call will proceed down the error path.
sim.SetInHoursCheck(func(name string, isQueue bool, time time.Time) (inOperation bool, err error) {
    return t.Hour() >= 8 && t.Hour() <= 18 && t.Weekday() != time.Sunday, nil
})
```

## Interacting with calls

```go
// Calls are started in their own go routine.
call, err := sim.StartCall(simulator.CallConfig{
    SourceNumber: "+447878987654",
    DestNumber: "+44113123456",
})

// To read spoken text, read from the caller output channel.
// The call will pause if this channel is not read.
prompt := <-call.Caller.O

// When prompted for input, write input one character at a time to the input channel.
// This write will block if the call is not taking caller input.
call.Caller.I <- '1'
call.Caller.I <- '#'

// For more detailed information about the call, register a lister on the event stream.
// All events will be sent to the provided channel. If the channel is blocked, the call will pause.
// Events are defined in the event package of this repository.
evtRcv := make(chan event.Event, 64)
call.Subscribe(evtRcv)
for {
    evt := <-evtRcv
    switch evt.Type() {
    case event.InvokeLambdaType:
        fmt.Println(evt.(event.InvokeLambdaEvent).ResponseJSON)
        break
    }
}

// Terminate the call when it is no longer needed.
call.Terminate()
```

## Testing

Automated testing is a big focus of this project. An assertion library is provided in the `flowtest` package. You can find examples of this in use in [simulator_test.go](./simulator_test.go)

### Attaching to a call

```go
func Test(t *testing.T) {
    call, err := sim.StartCall(simulator.CallConfig{
        SourceNumber: "+447878987654",
        DestNumber: "+44113123456",
    })

    // Create a new expect for each call.
    // The test helper will control the call. Do not directly interact with a call after this.
    expect := flowtest.New(t, call)

    // Begin assertions.
    expect.Prompt().ToContain("Welcome")

    // Simulate user input.
    expect.Caller().ToPress('1')
}
```

### Constructing assertions

#### Components

Assertions have this format:
```go
expect.Prompt().WithVoice("Joanna").WithSSML().ToContain("Hello")
```

Breaking this down:

* `expect` - The expect instance
* `Prompt()` - What we are making assertions about. One of `Caller`, `Prompt`, `Transfer`, `Lambda`, `Attributes`. These determine what valid assertions can follow.
* `WithVoice("Joanna")`, `WithSSML()` - a series of zero or more methods starting with `With`, that add conditions that must all be met for the assertion to pass.
* `ToContain("Hello")` - A method starting with `To` that executes the assertion.

If any of the `With` or `To` assertions fail, `fmt.Errorf` will be called with a summary of what failed.

#### Negation

Parts of the assertion can be negated by calling `Not()` before each assertion component that should be negated.

```go
    // Fails if greeting lambda is invoked. Other lambdas or no lambdas invoked is ok.
    expect().Lambda().WithARN("greeting").Not().ToBeInvoked()
    // A prompt must play. But it may not be read in the Joanna voice.
    expect().Prompt().Not().WithVoice("Joanna").ToPlay()
    // A lambda with greeter in its ARN must run, but may not return {"greeting": "hello"}
    expect().Lambda().Not().WithReturn("greeting", "hello").WithARN("greeter").ToSucceed()
```

#### Never

You can assert that something will not happen in any of the following testing by including `.Never()` somewhere in the assertion chain.

```go
    // We must never mix up SSML and Plaintext prompts.
    expect.Prompt().WithPlaintext().Never().ToContain("<speak>")
    expect.Prompt().WithSSML().Never().Not().ToContain("<speak>")
    // Any prompts must always be read in the Joanna voice.
    expect.Prompt().Never().Not().WithVoice("Joanna").ToPlay()
    // Any lambdas must be set to timeout after 6 seconds.
    expect.Lambda().Never().Not().WithTimeout(6 * time.Second).ToBeInvoked()
```

If the assertion matches at any time in the subsequent running of the call, `fmt.Errorf` will be called with a summary of what failed.

#### Ordering

There is an implicit assertion that the elements of the call you expect will occur in the order you expect them.
```go
// Implicitly assert that prompts occur in A->B->C order.
expect.Prompt().ToEqual("A")
expect.Prompt().ToEqual("B")
expect.Prompt().ToEqual("C")
```

You can temporarily disable this by adding `Unordered()` to the assertion chain.
```go
// A->B->C and A->C->B orderings will match.
// (B must still come after A but not necessarily before C).
expect.Prompt().ToEqual("A")
expect.Prompt().Unordered().ToEqual("B")
expect.Prompt().ToEqual("C")
```

### Simulating user actions

These are not assertions but rather allow the call to proceed by mocking caller behaviour. They may error if called when input is not required by the flow.
```go
expect := flowtest.New(t, call)

expect.Caller().ToPress('1') // Enter a single character.
expect.Caller().ToEnter("01234#") // Enter a sequence of characters.
expect.Caller().ToWaitForTimeout() // Wait for the menu to time out (actually takes zero time).
```

### `expect.Prompt()`

This context allows assertions on prompts read out by text-to-speach either as part of a Play Prompt block or from blocks like Get User Input.

```go

.WithSSML() // Prompt should be read as SSML
.WithPlaintext() // Prompt should be read as Plaintext
.WithVoice(voice string) // Prompt should be read in the given voice

.ToContain(text string) // Substring match for prompt content.
.ToEqual(text string) // Exact match for prompt content.
.ToPlay() // Matches any prompt.

```

### `expect.Lambda()`

This context allows assertions on invocation of lambdas.

```go
.WithTimeout(time time.Duration) // Lambda has a timeout configured.
.WithARN(arn string) // ARN of lambda contains the given string
.WithParameters(params map[string]string) // All of the given parmeters match those passed to the lambda.
.WithParameter(key string, value string) // This single parameter matches one of those passed to the lambda.
.WithReturns(params map[string]string) // All of the given key/values match those returned from the lambda.
.WithReturn(key string, value string) // This single key/value matches one of those returned from the lambda.

.ToBeInvoked() // A lambda is invoked.
.ToFail() // A lambda is invoked but returns an error.
.ToSucceed() // A lambda is invoked and it does not return an error.
```

### `expect.Attributes()`

This context allows assertions on user parameters being attached to the caller.

```go
.ToUpdateKey(key string) // Any value is set against the given key for the caller.
.ToUpdate(key string, value string) // The key and value given is set for the caller.
```

### `expect.Transfer()`

This context allows assertions about transfers to numbers, flows and queues.

```go
.ToQueue(named string) // The caller is transfered to a queue with the given name.
.ToFlow(named string) // The caller is transfered to a flow with the given name.
.ToNumber(tel string) // A caller is transfered to the given external number.
```

### Modularising tests

If you wish to modularise your tests while maintaining the fluent interface, you can use the following construct:

```go
// Define a set of assertions.
func getToMainMenu(expect *flowtest.Expect) {
    expect.Prompt().ToContain("Welcome")
    expect.Prompt().ToContain("Enter your account number")
    expect.Caller().ToEnter("123456")
    expect.Prompt().ToContain("Enter your date of birth")
    expect.Caller().ToEnter("03021991")
    expect.Prompt().ToContain("Main menu")
}

// Use the modular assertion.
expect.To(getToMainMenu)
expect.Caller().ToPress('2')
```

### Test Coverage

A tool is provided to track test coverage of your flows.

```go
// Create a single simulator and connect a coverage reporter.
sim := simulator.New()
coverage := flowtest.NewCoverageReporter(&sim)

for _, tC := range testCases {
    t.Run(tC.desc, func(t *testing.T) {
        call, err := sim.StartCall(tC.conf)
        if err != nil {
            t.Fatalf("unexpected error starting call: %v", err)
        }
        // For each call started, instruct the coverage reporter to track it.
        coverage.Track(call)
        // Then run tests once it's attached.
        expect := flowtest.New(t, call)
        expect.To(tC.assert)
        call.Terminate()
    })
}

// At the end, get the test coverage. This is measured as the fraction of links between blocks that at least one tracked call passed through.
fmt.Printf("Flow test coverage: %.0f%%\n", coverage.Coverage()*100)

// You can also get a full report of which branches were explored.
report := coverage.CoverageReport()

// A utility function is provided that can take this report and output a human-readable report that could be output to the terminal or written to a file.
str := flowtest.FormatCoverageReport(report, true)
```

### Debugger

Also included is a step-through debugger that support breakpoints, pausing, resuming and stepping through code. It presently serves no purpose. Note: attaching a debugger and an assertion helper to the same call may lead to undesirable outcomes.

```go
debugger := flowtest.NewDebugger(call)
debugger.SetBreakpoint("7eefafd6-402f-4759-967c-b017ef5f3969")
debugger.Wait()
```

## Example terminal integration

```go
func main () {
    // Start a call.
    call, err := sim.StartCall(simulator.CallConfig{
        SourceNumber: "+447878987654",
        DestNumber: "+44113123456",
    })
    if err != nil {
        panic(err)
    }

    // Take input from the terminal and pass it as input to the call.
    go func() {
        reader := bufio.NewReader(os.Stdin)
        for {
            char, _, err := reader.ReadRune()
            if err != nil {
                panic(err)
            }
            if char == '\n' {
                continue
            }
            call.Caller.I <- char
        }
    }()

    // Print output from the call back to the screen.
    for {
        if str, ok := <-call.Caller.O; ok {
            fmt.Println("> " + str)
        } else {
            break
        }
    }    
}
```
