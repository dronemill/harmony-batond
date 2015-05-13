package main

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/dronemill/harmony-client-go"
	"github.com/fsouza/go-dockerclient"
)

// Listener is the main app contianer
type Listener struct {
	// Harmony client
	Harmony *harmonyclient.Client

	// Docker client
	Dkr *docker.Client
}

// Listen for events
func (l *Listener) Listen() {
	log.WithField("dockerSock", config.DockerSock).Info("Opening Docker event listener")

	var c chan *docker.APIEvents
	c = make(chan *docker.APIEvents)

	l.Dkr.AddEventListener(c)

	for {
		v := <-c
		fmt.Printf("%+v\n", v)
	}
}
