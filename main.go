package main

import (
	"flag"
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/dronemill/harmony-client-go"
	"github.com/fsouza/go-dockerclient"
)

var batond *Batond
var listener *Listener
var stopped chan bool

func main() {
	stopped = make(chan bool)

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

	if !config.OneTime {
		go startEventListner()
	}

	batond = &Batond{
		Dkr:     dockerClient(),
		Harmony: harmonyClient(),
	}
	machine := batond.getMachine()

	log.WithField("machineID", machine.ID).
		WithField("name", config.Machine.Name).
		Info("Using machine by name")

	for _, cid := range machine.ContainerIDs {
		batond.checkContainerState(cid)
	}

	<-stopped
	log.Info("Shutting down batond...")
}

func stop() {
	stopped <- true
}

func startEventListner() {
	log.Info("Starting docker event listener")

	listener = &Listener{
		Dkr:     dockerClient(),
		Harmony: harmonyClient(),
	}

	listener.Listen()
}

// harmonyConnect will get a connected harmony client
func harmonyClient() *harmonyclient.Client {
	hconf := harmonyclient.Config{
		APIHost:      config.Harmony.API,
		APIVersion:   "v1",
		APIVerifySSL: config.Harmony.VerifySSL,
	}

	log.WithField("harmonyAPI", config.Harmony.API).Info("Attempting connection to HarmonyAPI")

	var err error
	h, err := harmonyclient.NewHarmonyClient(hconf)

	if err != nil {
		// TODO: maybe like dont bomb out here.. @pmccarren
		log.Fatalf("Failed connecting to the HarmonyAPI: %s", err.Error())
	}

	return h
}

func dockerClient() *docker.Client {
	log.WithField("dockerSock", config.DockerSock).Info("Attempting connection to Docker")

	endpoint := fmt.Sprintf("unix://%s", config.DockerSock)
	dkr, err := docker.NewClient(endpoint)
	if err != nil {
		log.WithField("dockerSock", config.DockerSock).
			WithField("error", err.Error()).
			Fatal("Failed creating new Docker client")
	}

	return dkr
}
