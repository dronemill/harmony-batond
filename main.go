package main

import (
	"flag"
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
)

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

	// get the hostname
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err.Error())
	}

	// check our sanity
	sanityCheck()

	log.WithField("hostname", hostname).Info("Starting batond")
}
