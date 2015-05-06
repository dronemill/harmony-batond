package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/dronemill/harmony-client-go"
)

// Batond is the main app contianer
type Batond struct {
	// hamrony client
	Harmony *harmonyclient.Client
}

// maestroConnect will get a connected maestro client
func (b *Batond) harmonyConnect() error {
	hconf := harmonyclient.Config{
		APIHost:      config.Harmony.API,
		APIVersion:   "v1",
		APIVerifySSL: config.Harmony.VerifySSL,
	}

	log.WithField("harmonyAPI", config.Harmony.API).Info("Attempting connection to HarmonyAPU")

	var err error
	b.Harmony, err = harmonyclient.NewHarmonyClient(hconf)

	if err != nil {
		// FIXME: maybe like dont bomb out here.. @pmccarren
		log.Fatalf("Failed connecting to the HarmonyAPI: %s", err.Error())
	}

	return nil
}

// getMachine will get the Harmony Machine
func (b *Batond) getMachine() *harmonyclient.Machine {
	// check if we already have a MachineID
	machine := b.getMachineByName()
	if machine != nil {
		return machine
	}

	return b.createMachine()
}

// getMachineByName will get a Harmony Machine resource by name
func (b *Batond) getMachineByName() *harmonyclient.Machine {
	machine, err := b.Harmony.MachineByName(config.Machine.Name)
	if err != nil {
		// FIXME: need to get some better error handling in here.. @pmccarren
		log.WithField("name", config.Machine.Name).
			WithField("error", err.Error()).
			Fatalf("Failed fetching machine by name")
	}

	return machine
}

// createMachine will create a Harmony Machine resource
func (b *Batond) createMachine() *harmonyclient.Machine {

	machine := &harmonyclient.Machine{
		Name:     config.Machine.Name,
		Hostname: config.Machine.Hostname,
	}

	machineResource, err := b.Harmony.MachinesAdd(machine)

	if err != nil {
		log.WithField("error", err.Error()).
			Fatalf("Failed creating machine")
	}

	return machineResource
}
