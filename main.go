package main

import (
	"flag"
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
)

func main() {
	flag.Parse()
	if printVersion {
		fmt.Printf("confd %s\n", Version)
		os.Exit(0)
	}
	if err := initConfig(); err != nil {
		log.Fatal(err.Error())
	}
	log.Info("Starting batond")

	sanityCheck()
}
