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
		harmonyAPI string
	}

	configFile        = ""
	defaultConfigFile = "config.toml"
	printVersion      bool
)

func init() {
	flag.StringVar(&configFile, "configFile", "", "the config file")
	flag.BoolVar(&printVersion, "version", false, "print version and exit")
	flag.StringVar(&c.logLevel, "logLevel", "", "the level of messages to log")

	flag.StringVar(&c.harmonyAPI, "harmonyAPI", "http://harmony.dev:4774", "the url to the Harmony API")
}

type Config struct {
	LogLevel string `toml:LogLevel`

	HarmonyAPI string `toml:HarmonyAPI`
}

func initConfig() error {
	if configFile == "" {
		if _, err := os.Stat(defaultConfigFile); !os.IsNotExist(err) {
			configFile = defaultConfigFile
		}
	}

	// Set defaults.
	config = Config{
		LogLevel:   "info",
		HarmonyAPI: "http://harmony.dev:4774",
	}

	// Update config from the TOML configuration file.
	if configFile == "" {
		log.Warning("Skipping config file parsing")
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

	case "harmonyAPI":
		config.HarmonyAPI = c.harmonyAPI
	}
}
