package main

import (
	"testing"
)

// Test configuration loading and parsing
// using the default config

func TestLoadConfigs(t *testing.T) {

	config, err := loadConfig("../etc/alicelg/alice.example.conf")
	if err != nil {
		t.Error("Could not load test config:", err)
	}

	if config.Server.Listen == "" {
		t.Error("Listen string not present.")
	}

	if len(config.Ui.RoutesColumns) == 0 {
		t.Error("Route columns settings missing")
	}

	if len(config.Ui.RoutesRejections.Reasons) == 0 {
		t.Error("Rejection reasons missing")
	}
}

func TestSourceConfigDefaultsOverride(t *testing.T) {

	config, err := loadConfig("../etc/alicelg/alice.example.conf")
	if err != nil {
		t.Error("Could not load test config:", err)
	}

	// Get sources

	rs1 := config.Sources[0]
	rs2 := config.Sources[1]

	// Source 1 should be on default time
	// Source 2 should have an override
	// For now it should be sufficient to test if
	// the serverTime(rs1) != serverTime(rs2)
	if rs1.Birdwatcher.ServerTime == rs2.Birdwatcher.ServerTime {
		t.Error("Server times should be different between",
			"source 1 and 2 in example configuration",
			"(alice.example.conf)")
	}

}
