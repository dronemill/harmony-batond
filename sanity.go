package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
)

// sanityCheck checks that the system is in a state to be conducted
func sanityCheck() {
	log.Info("Running a sanity check")

	sanityCheckDockerSockExists()
	checkMachineHostname()
	checkMachineName()
}

func sanityCheckDockerSockExists() {
	// equivalent to Python's `if not os.path.exists(filename)`
	if _, err := os.Stat(config.DockerSock); os.IsNotExist(err) {
		log.Fatalf("Docker Sock does not exist!: '%s' file not found", config.DockerSock)
		return
	}

	log.Debug("Sanity of Docker Socket: sane")
}

func checkMachineHostname() {
	if config.Machine.Hostname == "" {
		log.Fatal("Machine Hostname not set")
	}

	log.Debug("Sanity of Machine Hostname: sane")
}

func checkMachineName() {
	if config.Machine.Name == "" {
		log.Fatal("Machine Name not set")
	}

	log.Debug("Sanity of Machine Name: sane")
}
