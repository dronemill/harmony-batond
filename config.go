package main

import (
	"flag"
	"io/ioutil"
	"os"

	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
)

var (
	config Config // holds the global config

	c struct {
		logLevel   string
		oneTime    bool
		dockerSock string
		harmony    struct {
			api       string
			verifyssl bool
		}
		maestro struct {
			host      string
			eventPort int // the EventSocket port
		}
		machine struct {
			hostname string
			name     string
		}
	}

	configFile        = ""
	defaultConfigFile = "config.toml"
	printVersion      bool
)

func init() {
	flag.StringVar(&configFile, "configFile", "", "the config file")
	flag.BoolVar(&printVersion, "version", false, "print version and exit")

	flag.StringVar(&c.logLevel, "logLevel", "", "the level of messages to log")
	flag.BoolVar(&c.oneTime, "oneTime", false, "run once and exit (no event watching)")
	flag.StringVar(&c.dockerSock, "dockerSock", "/tmp/docker.sock", "Docker Daemon control socket")

	flag.StringVar(&c.harmony.api, "harmony.api", "http://harmony.dev:4774", "the url to the Harmony API")
	flag.BoolVar(&c.harmony.verifyssl, "harmony.verifyssl", true, "verify ssl connections to the harmony api")

	flag.StringVar(&c.maestro.host, "maestro.host", "harmony.dev", "the ip/hostname of the maestro")
	flag.IntVar(&c.maestro.eventPort, "maestro.eventPort", 4775, "the port of the maestro's EventSocket server")

	flag.StringVar(&c.machine.hostname, "machine.hostname", "", "Harmony machine name")
	flag.StringVar(&c.machine.name, "machine.name", "", "Harmony machine name")
}

// Config is the main config type
type Config struct {
	// LogLevel main application loggin level
	LogLevel string `toml:"LogLevel"`

	// OneTime is this a single run
	OneTime bool `toml:"OneTime"`

	// Harmony is the main Harmony config
	Harmony HarmonyConfig `toml:"Harmony"`

	// Maestro contains the configuration to the maestro
	Maestro MaestroConfig `toml:"Maestro"`

	// DockerSock is the path to the Docker Daemon control socket
	DockerSock string `toml:"DockerSock"`

	// Machine holds the harmony machine configuration
	Machine MachineConfig `toml:"Machine"`
}

// HarmonyConfig is the main Harmony config
type HarmonyConfig struct {
	// API url to the Harmony API
	API string `toml:"API"`

	// VerifySSL is wether ot not we are to verify the secure Harmony API connections
	VerifySSL bool `toml:"VerifySSL"`
}

// MaestroConfig contains the configuration to the maestro
type MaestroConfig struct {
	// Host is the ip/hostname of the maestro
	Host string `toml:"Host"`

	// EventPort the port of the maestro's EventSocket server
	EventPort int `toml:"EventPort"`
}

// MachineConfig holds the harmony machine configuration
type MachineConfig struct {
	// Hostname is the Harmony Machine Hostname
	Hostname string `toml:"Hostname"`

	// Name is the Harmony Machine Name
	Name string `toml:"Name"`
}

func initConfig() error {
	if configFile == "" {
		if _, err := os.Stat(defaultConfigFile); !os.IsNotExist(err) {
			configFile = defaultConfigFile
		}
	}

	// get the hostname
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err.Error())
	}

	// Set defaults.
	config = Config{
		LogLevel:   "info",
		OneTime:    false,
		DockerSock: "/tmp/docker.sock",
		Harmony: HarmonyConfig{
			API:       "http://harmony.dev:4774",
			VerifySSL: true,
		},
		Maestro: MaestroConfig{
			Host:      "harmony.dev",
			EventPort: 4775,
		},
		Machine: MachineConfig{
			Hostname: hostname,
			Name:     "",
		},
	}

	// Update config from the TOML configuration file.
	if configFile == "" {
		log.Info("Skipping config file parsing")
	} else {
		log.WithField("file", configFile).Info("Loading config")

		configBytes, err := ioutil.ReadFile(configFile)
		if err != nil {
			return err
		}
		_, err = toml.Decode(string(configBytes), &config)
		if err != nil {
			return err
		}
	}
	// Update config from commandline flags.
	processFlags()

	if config.LogLevel != "" {
		LogSetLevel(config.LogLevel)
	}

	return nil
}

func processFlags() {
	flag.Visit(setConfigFromFlag)
}

func setConfigFromFlag(f *flag.Flag) {
	switch f.Name {
	case "logLevel":
		config.LogLevel = c.logLevel
	case "oneTime":
		config.OneTime = c.oneTime

	case "dockerSock":
		config.DockerSock = c.dockerSock

	case "harmony.api":
		config.Harmony.API = c.harmony.api
	case "harmony.verifyssl":
		config.Harmony.VerifySSL = c.harmony.verifyssl

	case "maestro.host":
		config.Maestro.Host = c.maestro.host
	case "maestro.eventPort":
		config.Maestro.EventPort = c.maestro.eventPort

	case "machine.hostname":
		config.Machine.Hostname = c.machine.hostname
	case "machine.name":
		config.Machine.Name = c.machine.name
	}
}
