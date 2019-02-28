package main

import (
	"github.com/d2r2/go-logger"
	"github.com/jakekeeys/hatchery/internal/monitor"
	"github.com/jakekeeys/hatchery/internal/relay"
	"github.com/jakekeeys/hatchery/internal/sensor"
	"github.com/jakekeeys/hatchery/internal/server"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	err := logger.ChangePackageLogLevel("dht", logger.FatalLevel)
	if err != nil {
		logrus.Panic(err)
	}

	logrus.SetLevel(logrus.DebugLevel)

	sensor := sensor.New()
	defer sensor.Stop()
	sensor.Start()

	relay := relay.New()
	defer relay.Close()

	monitor := monitor.New(sensor, relay)
	defer monitor.Stop()
	monitor.Start()

	server := server.New(sensor, relay)
	defer server.Stop()
	server.Start()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
}
