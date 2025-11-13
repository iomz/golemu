//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package config

import (
	"net"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	// Version is the current version
	Version = "0.5.1"

	// App is the kingpin application
	App = kingpin.New("golemu", "A mock LLRP-based logical reader emulator for RFID Tags.")

	// Global flags
	Debug              = App.Flag("debug", "Enable debug mode.").Short('v').Default("false").Bool()
	InitialMessageID   = App.Flag("initialMessageID", "The initial messageID to start from.").Default("1000").Int()
	InitialKeepaliveID = App.Flag("initialKeepaliveID", "The initial keepaliveID to start from.").Default("80000").Int()
	IP                 = App.Flag("ip", "LLRP listening address.").Short('a').Default("0.0.0.0").IP()
	KeepaliveInterval  = App.Flag("keepalive", "LLRP Keepalive interval.").Short('k').Default("0").Int()
	Port               = App.Flag("port", "LLRP listening port.").Short('p').Default("5084").Int()
	PDU                = App.Flag("pdu", "The maximum size of LLRP PDU.").Short('m').Default("1500").Int()
	ReportInterval     = App.Flag("reportInterval", "The interval of ROAccessReport in ms. Pseudo ROReport spec option.").Short('i').Default("10000").Int()

	// Client mode
	Client = App.Command("client", "Run as an LLRP client; connect to an LLRP server and receive events (test-only).")

	// Server mode
	Server  = App.Command("server", "Run as an LLRP tag stream server.")
	APIPort = Server.Flag("apiPort", "The port for the API endpoint.").Default("3000").Int()
	File    = Server.Flag("file", "The file containing Tag data.").Short('f').Default("tags.gob").String()

	// Simulator mode
	Simulator     = App.Command("simulator", "Run in the simulator mode.")
	SimulationDir = Simulator.Arg("simulationDir", "The directory contains tags for each event cycle.").Required().String()
)

// Config holds all application configuration values parsed from command-line flags.
// It provides a structured way to access configuration throughout the application.
type Config struct {
	Debug              bool
	InitialMessageID   int
	InitialKeepaliveID int
	IP                 net.IP
	KeepaliveInterval  int
	Port               int
	PDU                int
	ReportInterval     int
	APIPort            int
	File               string
	SimulationDir      string
}

// GetConfig returns the current application configuration parsed from command-line flags.
// It should be called after kingpin.Parse() to ensure all flags are populated.
//
// Returns a Config struct containing all configuration values.
func GetConfig() *Config {
	return &Config{
		Debug:              *Debug,
		InitialMessageID:   *InitialMessageID,
		InitialKeepaliveID: *InitialKeepaliveID,
		IP:                 *IP,
		KeepaliveInterval:  *KeepaliveInterval,
		Port:               *Port,
		PDU:                *PDU,
		ReportInterval:     *ReportInterval,
		APIPort:            *APIPort,
		File:               *File,
		SimulationDir:      *SimulationDir,
	}
}
