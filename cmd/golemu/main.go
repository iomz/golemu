//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/iomz/golemu/config"
	"github.com/iomz/golemu/connection"
	"github.com/iomz/golemu/server"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	// Set version
	config.App.Version(config.Version)
	parse := kingpin.MustParse(config.App.Parse(os.Args[1:]))

	// Set up logrus
	log.SetLevel(log.InfoLevel)

	cfg := config.GetConfig()
	if cfg.Debug {
		gin.SetMode(gin.DebugMode)
		log.SetLevel(log.DebugLevel)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	switch parse {
	case config.Client.FullCommand():
		client := connection.NewClient(cfg.IP.String(), cfg.Port)
		os.Exit(client.Run())
	case config.Server.FullCommand():
		srv := server.NewServer(
			cfg.IP.String(),
			cfg.Port,
			cfg.APIPort,
			cfg.PDU,
			cfg.ReportInterval,
			cfg.KeepaliveInterval,
			cfg.InitialMessageID,
			cfg.File,
		)
		os.Exit(srv.Run())
	case config.Simulator.FullCommand():
		sim := connection.NewSimulator(
			cfg.IP.String(),
			cfg.Port,
			cfg.PDU,
			cfg.ReportInterval,
			cfg.SimulationDir,
			cfg.InitialMessageID,
		)
		os.Exit(sim.Run())
	}
}
