package main

import (
	simulator "amazon-connect-simulator"
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	checkaccountnum "lambdas/checkaccountnum/handler"
	checkdob "lambdas/checkdob/handler"
	checkpass "lambdas/checkpass/handler"
	genpass "lambdas/genpass/handler"
	greeting "lambdas/greeting/handler"
	"os"
	"reflect"
	"time"
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

	fng := greeting.New()
	fncan := checkaccountnum.New(func(num string) (string, bool, error) {
		return "CUSTOMERID", num == "12345678", nil
	})
	fncd := checkdob.New(func(customerID string) (time.Time, error) {
		return time.Date(1991, 04, 04, 0, 0, 0, 0, time.Local), nil
	})
	fngp := genpass.New(func(customerID string) (string, []int, error) {
		return "chalID", []int{0, 2}, nil
	})
	fncp := checkpass.New(func(customerID string, chalID string, dig []int) (bool, error) {
		return reflect.DeepEqual(dig, []int{1, 3}), nil
	})

	sim.RegisterLambda("Welcome", fng.Handle)
	sim.RegisterLambda("CheckAccountNumber", fncan.Handle)
	sim.RegisterLambda("CheckDoB", fncd.Handle)
	sim.RegisterLambda("GeneratePassChallenge", fngp.Handle)
	sim.RegisterLambda("CheckPassChallenge", fncp.Handle)

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
