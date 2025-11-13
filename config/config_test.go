//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package config

import (
	"testing"
)

func TestGetConfig(t *testing.T) {
	cfg := GetConfig()

	if cfg == nil {
		t.Fatal("GetConfig returned nil")
	}

	// Note: GetConfig reads from kingpin flags which are only populated
	// after parsing command line arguments. In tests, these will be zero values
	// unless we parse args first. This test just verifies the function doesn't panic.
	_ = cfg.IP
	_ = cfg.Port
	_ = cfg.PDU
}

func TestConfig_Structure(t *testing.T) {
	cfg := GetConfig()

	if cfg == nil {
		t.Fatal("GetConfig returned nil")
	}

	// Verify structure exists (values depend on kingpin parsing)
	_ = cfg.Debug
	_ = cfg.InitialMessageID
	_ = cfg.InitialKeepaliveID
	_ = cfg.IP
	_ = cfg.KeepaliveInterval
	_ = cfg.Port
	_ = cfg.PDU
	_ = cfg.ReportInterval
	_ = cfg.APIPort
	_ = cfg.File
	_ = cfg.SimulationDir
}

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}
}
