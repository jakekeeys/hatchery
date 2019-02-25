package monitor

import (
	"context"
	"github.com/jakekeeys/hatchery/internal/relay"
	"github.com/jakekeeys/hatchery/internal/sensor"
	"github.com/sirupsen/logrus"
	"time"
)

type Monitor struct {
	ctx context.Context
	sensor  *sensor.Sensor
	relay   *relay.Relay
}

func New(sensor *sensor.Sensor, relay *relay.Relay) *Monitor {
	return &Monitor{
		ctx: context.Background(),
		sensor: sensor,
		relay: relay,
	}
}


func (m Monitor) Start() {
	go func() {
		for {
			select {
			case <-time.After(time.Minute*1):
				err := m.poll()
				if err != nil {
					logrus.Warn(err)
				}
			case <-m.ctx.Done():
				return
			}
		}
	}()
}

func (m Monitor) Stop() {
	m.ctx.Done()
}

func (m Monitor) poll() error {
	_, h := m.sensor.GetTemperatureAndHumidity()
	logrus.Infof("humidity %f", h)

	if h < 45 {
		logrus.Info("turning pump on")
		m.relay.SetOn()
		time.Sleep(time.Second*5)
		logrus.Info("turning pump off")
		m.relay.SetOff()
	}

	return nil
}