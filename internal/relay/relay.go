package relay

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
	"github.com/stianeikeland/go-rpio"
	"sync"
)

const (
	relaySwitchPin = 3
	relayOnState   = 1
	relayOffState  = 0
)

type Relay struct {
	pin        *rpio.Pin
	mtx        *sync.Mutex
	on         *bool
	relayGauge prometheus.Gauge
}

func New() *Relay {
	relayGauge := promauto.NewGauge(prometheus.GaugeOpts{
		Name: "relay",
	})

	err := rpio.Open()
	if err != nil {
		logrus.Panic(err)
	}

	pin := rpio.Pin(relaySwitchPin)
	pin.Output()
	pin.High()
	b := false

	return &Relay{
		pin:        &pin,
		mtx:        &sync.Mutex{},
		on:         &b,
		relayGauge: relayGauge,
	}
}

func (r Relay) Close() {
	r.pin.High()
	err := rpio.Close()
	if err != nil {
		logrus.Warn(err)
	}
}

func (r Relay) SetOn() {
	r.pin.Low()

	b := true
	r.mtx.Lock()
	*r.on = b
	r.mtx.Unlock()

	r.relayGauge.Set(relayOnState)
}

func (r Relay) SetOff() {
	r.pin.High()

	b := false
	r.mtx.Lock()
	*r.on = b
	r.mtx.Unlock()

	r.relayGauge.Set(relayOffState)
}

func (r Relay) IsOn() bool {
	var b bool
	r.mtx.Lock()
	b = *r.on
	r.mtx.Unlock()

	return b
}
