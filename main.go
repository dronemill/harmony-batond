package main

import (
	"flag"
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/dronemill/eventsocket-client-go"
	"github.com/dronemill/harmony-client-go"
	"github.com/fsouza/go-dockerclient"
)

var (
	batond   *Batond
	maestro  *Maestro
	listener *Listener
	machine  *harmonyclient.Machine
	stopped  chan bool
)

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
	machine = batond.getMachine()

	go startMaestro()

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

func startMaestro() {
	maestro = &Maestro{
		Harmony: harmonyClient(),
		Portal:  maestroPortal(),
	}

	maestro.portalEmitExistance()
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

func maestroPortal() *eventsocketclient.Client {
	log.WithField("maestro", config.Maestro.Host).WithField("portalPort", config.Maestro.PortalPort).Info("Opening connection to Maestro Portal")

	portal, err := eventsocketclient.NewClient(fmt.Sprintf("%s:%d", config.Maestro.Host, config.Maestro.PortalPort))
	if err != nil {
		log.WithField("maestro", config.Maestro.Host).
			WithField("portalPort", config.Maestro.PortalPort).
			WithField("error", err.Error()).
			Fatal("Failed connecting to Maestro Portal")
	}

	log.WithField("maestro", config.Maestro.Host).WithField("portalPort", config.Maestro.PortalPort).Debug("DialingWS")

	if err := portal.DialWs(); err != nil {
		log.WithField("maestro", config.Maestro.Host).
			WithField("portalPort", config.Maestro.PortalPort).
			WithField("esID", portal.Id).
			WithField("error", err.Error()).
			Fatal("Failed dialingWS")
	}

	log.WithField("maestro", config.Maestro.Host).WithField("portalPort", config.Maestro.PortalPort).WithField("esID", portal.Id).Info("Successfully DialedWS")

	portal.SetMaxMessageSize(5242880)    // 5MB
	log.WithFields(log.Fields{"maestro": config.Maestro.Host,
		"portalPort": config.Maestro.PortalPort,
		"esID":       portal.Id,
		"size":       5242880,
	}).Debug("Set max message size")

	log.WithField("maestro", config.Maestro.Host).WithField("portalPort", config.Maestro.PortalPort).WithField("esID", portal.Id).Info("Successfully connected to the Maestro Portal")

	return portal
}
