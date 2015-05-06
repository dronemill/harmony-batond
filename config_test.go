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
		HarmonyAPI: "http://harmony.dev:4774",
		DockerSock: "/tmp/docker.sock",
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
