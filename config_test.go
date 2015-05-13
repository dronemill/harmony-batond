package main

import (
	"log"
	"os"
	"reflect"
	"testing"
)

func TestInitConfigDefaultConfig(t *testing.T) {
	// get the hostname
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err.Error())
	}

	LogSetLevel("warning")
	want := Config{
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
	if err := initConfig(); err != nil {
		t.Errorf(err.Error())
	}
	if !reflect.DeepEqual(want, config) {
		t.Errorf("initConfig() = %v, want %v", config, want)
	}
}
