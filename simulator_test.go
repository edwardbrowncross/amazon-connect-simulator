package simulator

// var testJSON = `{
// "parameters": [
// 	{
// 		"name": "Text",
// 		"value": "Please enter your date of birth.",
// 		"namespace": null
// 	},
// 	{
// 		"name": "TextToSpeechType",
// 		"value": "text"
// 	},
// 	{
// 		"name": "Timeout",
// 		"value": "5"
// 	},
// 	{
// 		"name": "MaxDigits",
// 		"value": 8
// 	},
// 	{
// 		"name": "EncryptEntry",
// 		"value": true
// 	},
// 	{
// 		"name": "Parameter",
// 		"value": "testValue",
// 		"key": "testKey"
// 	},
// 	{
// 		"name": "Parameter",
// 		"value": "testValue2",
// 		"key": "testKey2"
// 	}
// ]}`

// func TestUnmarshal(t *testing.T) {
// 	s := NewCallSimulator()
// 	m := Module{}
// 	err := json.Unmarshal([]byte(testJSON), &m)
// 	if err != nil {
// 		t.Fatalf("error perparing parameters: %v", err)
// 	}
// 	type someText string
// 	into := struct {
// 		Text         someText
// 		Timeout      string
// 		MaxDigits    int
// 		EncryptEntry bool
// 		Parameter    []KeyValue
// 	}{}
// 	err = s.unmarshalParameters(m.Parameters, &into)
// 	if err != nil {
// 		t.Errorf("unexpected error: %v", err)
// 	}
// 	if into.Text != "Please enter your date of birth." {
// 		t.Errorf("incorrect Text")
// 	}
// 	if into.Timeout != "5" {
// 		t.Errorf("incorrect Timeout")
// 	}
// 	if into.MaxDigits != 8 {
// 		t.Errorf("incorrect MaxDigits")
// 	}
// 	if into.EncryptEntry != true {
// 		t.Errorf("incorrect EncryptEntry")
// 	}
// 	if into.Parameter[0].k != "testKey" {
// 		t.Errorf("incorrect Parameter 0 k")
// 	}
// 	if into.Parameter[0].v != "testValue" {
// 		t.Errorf("incorrect Parameter 0 v")
// 	}
// 	if into.Parameter[1].k != "testKey2" {
// 		t.Errorf("incorrect Parameter 1 k")
// 	}
// 	if into.Parameter[1].v != "testValue2" {
// 		t.Errorf("incorrect Parameter 1 v")
// 	}
// 	fmt.Println(fmt.Sprintf("%v", into))
// }

// func TestInvoke(t *testing.T) {
// 	mod := invokeExternalResource{}
// 	type Input struct {
// 		A string `json:"a"`
// 		B string `json:"b"`
// 		C string `json:"c"`
// 	}
// 	type Output struct {
// 		X string `json:"x"`
// 		Y string `json:"y"`
// 		Z string `json:"z"`
// 	}
// 	test := func(ctx context.Context, in Input) (out Output, err error) {
// 		expIn := Input{
// 			A: "foo",
// 			B: "bar",
// 		}
// 		if !reflect.DeepEqual(expIn, in) {
// 			t.Errorf("expected '%v' but got '%v'", expIn, in)
// 		}
// 		return Output{
// 			X: "X",
// 			Y: "YY",
// 			Z: "ZZZ",
// 		}, nil
// 	}
// 	res, err := mod.invoke(test, `{"a":"foo","b":"bar","d":"baz"}`)
// 	if err != nil {
// 		t.Fatalf("unexpeced error: %v", err)
// 	}
// 	expRes := `{"x":"X","y":"YY","z":"ZZZ"}`
// 	if res != expRes {
// 		t.Errorf("expected '%s' but got '%s'", expRes, res)
// 	}
// }
