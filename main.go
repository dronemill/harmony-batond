package main

import (
	"flag"
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
)

var batond *Batond

func main() {
	// parse cli flags
	flag.Parse()

	// if we are just supposed to print the version, then do so
	if printVersion {
		fmt.Printf("confd %s\n", Version)
		os.Exit(0)
	}

	// load configuration
	if err := initConfig(); err != nil {
		log.Fatal(err.Error())
	}

	// check our sanity
	sanityCheck()

	log.WithFields(log.Fields{"machineName": config.Machine.Name, "machineHostname": config.Machine.Hostname}).Info("Starting batond")

	batond = &Batond{}
	batond.harmonyConnect()
	machine := batond.getMachine()

	log.WithField("machineID", machine.ID).
		WithField("name", config.Machine.Name).
		Info("Using machine by name")
}
