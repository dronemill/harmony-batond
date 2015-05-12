package main

import (
	"flag"
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
)

var batond *Batond
var dkr *docker.Client

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

	var err error
	endpoint := fmt.Sprintf("unix://%s", config.DockerSock)
	dkr, err = docker.NewClient(endpoint)
	if err != nil {
		log.WithField("socket", config.DockerSock).
			WithField("error", err.Error()).
			Fatal("Failed creating new Docker client")
	}

	batond = &Batond{}
	batond.harmonyConnect()
	machine := batond.getMachine()

	log.WithField("machineID", machine.ID).
		WithField("name", config.Machine.Name).
		Info("Using machine by name")

	for _, cid := range machine.ContainerIDs {
		batond.checkContainerState(cid)
	}
}
