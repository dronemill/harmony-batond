package main

import (
	"reflect"
	"testing"
)

func TestInitConfigDefaultConfig(t *testing.T) {
	LogSetLevel("warning")
	want := Config{
		LogLevel:   "info",
		HarmonyAPI: "http://harmony.dev:4774",
		DockerSock: "/tmp/docker.sock",
	}
	if err := initConfig(); err != nil {
		t.Errorf(err.Error())
	}
	if !reflect.DeepEqual(want, config) {
		t.Errorf("initConfig() = %v, want %v", config, want)
	}
}
