package main

import (
	"testing"

	log "github.com/Sirupsen/logrus"
)

func TestLogSetTag(t *testing.T) {
	LogSetTag("SuperAwesomeTag")

	if logTag != "SuperAwesomeTag" {
		t.Errorf("LogSetTag: did not set tag to 'SuperAwesomeTag'")
	}
}

func TestLogSetLevel(t *testing.T) {
	LogSetLevel("info")
	if log.GetLevel() != log.InfoLevel {
		t.Errorf("LogSetLevel: failed setting log level to 'info'")
	}

	LogSetLevel("warning")
	if log.GetLevel() != log.WarnLevel {
		t.Errorf("LogSetLevel: failed setting log level to 'warn'")
	}

	LogSetLevel(config.LogLevel)
}
