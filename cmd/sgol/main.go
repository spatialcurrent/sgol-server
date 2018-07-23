package main

import (
	//"bufio"
	//"bytes"
	//"flag"
	"fmt"
	//"io/ioutil"
	//"net/http"
	//"net/url"
	"os"
	//"strings"
	//"text/template"
	"time"
)

import (
//"github.com/hashicorp/hcl"
//"github.com/mattn/go-colorable"
//"github.com/pkg/errors"
//"github.com/sirupsen/logrus"
)

import (
	"github.com/spatialcurrent/sgol-server/sgol"
)

var SGOL_SERVER_VERSION = "0.0.1"

func main() {

	start := time.Now()

	if len(os.Args) < 2 {
		fmt.Println("Missing subcommand")
		fmt.Println("Run \"sgol help\" for command line options.")
		os.Exit(1)
	}

	config, err := sgol.LoadConfig(os.Getenv("SGOL_CONFIG_PATH"), true)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cmd := os.Args[1]

	commands := map[string]sgol.Command{}
	//commands["help"] = &sgol.HelpCommand{}
	commands["serve"] = sgol.NewServeCommand(config)
	commands["schema"] = sgol.NewSchemaCommand(config)
	commands["validate"] = sgol.NewValidateCommand(config)
	commands["operations"] = sgol.NewOperationsCommand(config)
	commands["filters"] = sgol.NewFiltersCommand(config)
	commands["queries"] = sgol.NewQueriesCommand(config)
	commands["formats"] = sgol.NewFormatsCommand(config)
	commands["exec"] = sgol.NewExecCommand(config)
	commands["add"] = sgol.NewAddCommand(config)
	//commands["lint"] = &sgol.LintCommand{}

	if cmd == "help" || cmd == "--help" || cmd == "-h" || cmd == "-help" {
		fmt.Println("SGOL Server - Version " + SGOL_SERVER_VERSION)
		fmt.Println("Commands:")
		for x, _ := range commands {
			fmt.Println("\t- sgol " + x)
		}
		os.Exit(0)
	}

	if _, ok := commands[cmd]; !ok {
		fmt.Println("Error: Unknown subcommand")
		fmt.Println("Run \"sgol help\" for command line options.")
		os.Exit(1)
	}

	err = commands[cmd].Parse(os.Args[2:])
	if err != nil {
		fmt.Println(err)
		fmt.Println("Run \"sgol " + cmd + " -help\" for command line options.")
		os.Exit(1)
	}

	err = commands[cmd].Run(start, SGOL_SERVER_VERSION)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
