package monitor

import (
	"context"
	"github.com/jakekeeys/hatchery/internal/relay"
	"github.com/jakekeeys/hatchery/internal/sensor"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	upperHumidityThreshold = 46.0
	lowerHumidityThreshold = 44.0
	targetHumidity         = 45.0

	pumpPulseDurationSeconds = 10

	monitorPollIntervalSeconds = 15
	deltaDurationSeconds       = 30
)

type Monitor struct {
	ctx    context.Context
	sensor *sensor.Sensor
	relay  *relay.Relay
}

func New(sensor *sensor.Sensor, relay *relay.Relay) *Monitor {
	return &Monitor{
		ctx:    context.Background(),
		sensor: sensor,
		relay:  relay,
	}
}

func (m Monitor) Start() {
	go func() {
		for {
			select {
			case <-time.After(time.Second * monitorPollIntervalSeconds):
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
	t, h := m.sensor.GetTemperatureAndHumidity()
	logrus.Debugf("current temperature: %f, humidity: %f", t, h)

	if h > lowerHumidityThreshold && h < upperHumidityThreshold {
		return nil
	}

	_, dh := m.getSensorDeltas(t, h)

	if h < targetHumidity && dh <= 0 {
		logrus.Debug("humidity less than target and trending down, activating pump")
		m.pulsePump()
	}

	return nil
}

func (m Monitor) getSensorDeltas(ct, ch float32) (float32, float32) {
	time.Sleep(time.Second * deltaDurationSeconds)
	t, h := m.sensor.GetTemperatureAndHumidity()

	dt := t - ct
	dh := h - ch
	logrus.Debugf("delta temperature %f, humidity %f", dt, dh)

	return dt, dh
}

func (m Monitor) pulsePump() {
	logrus.Debug("turning pump on")
	m.relay.SetOn()
	time.Sleep(time.Second * pumpPulseDurationSeconds)
	logrus.Debug("turning pump off")
	m.relay.SetOff()
}
