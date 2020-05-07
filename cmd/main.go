package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	simulator "github.com/edwardbrowncross/amazon-connect-simulator"
)

var files = []string{
	"Account Number.json",
	"Date of Birth.json",
	"Main Menu.json",
	"Passcode.json",
	"Start & Setup.json",
	"Welcome.json",
}

func main() {
	sim := simulator.NewCallSimulator()

	for _, f := range files {
		file, err := ioutil.ReadFile("./flows/" + f)
		if err != nil {
			panic(err)
		}
		flow := simulator.Flow{}
		err = json.Unmarshal(file, &flow)
		if err != nil {
			panic(err)
		}
		sim.LoadFlow(flow)
	}

	go func() {
		for {
			select {
			case str := <-sim.O:
				fmt.Println(str)
			}
		}
	}()

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
			sim.I <- char
		}
	}()

	err := sim.StartCall("Start & Setup")
	if err != nil {
		panic(err)
	}

	time.Sleep(1 * time.Second)
}
