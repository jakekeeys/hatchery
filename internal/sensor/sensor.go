package sensor

import (
	"context"
	"github.com/d2r2/go-dht"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

const (
	sensorPollIntervalSeconds = 1
	sensorDataPin             = 4
	sensorRetryLimit          = 10
)

type Sensor struct {
	ctx        context.Context
	mtx        *sync.Mutex
	temp       *float32
	humid      *float32
	tempGauge  prometheus.Gauge
	humidGauge prometheus.Gauge
}

func New() *Sensor {
	tempGauge := promauto.NewGauge(prometheus.GaugeOpts{
		Name: "temperature",
	})
	humidGauge := promauto.NewGauge(prometheus.GaugeOpts{
		Name: "humidity",
	})

	var temp, humid float32
	return &Sensor{
		ctx:        context.Background(),
		mtx:        &sync.Mutex{},
		temp:       &temp,
		humid:      &humid,
		tempGauge:  tempGauge,
		humidGauge: humidGauge,
	}
}

func (s Sensor) Start() {
	go func() {
		for {
			select {
			case <-time.After(time.Second * sensorPollIntervalSeconds):
				err := s.update()
				if err != nil {
					logrus.Warn(err)
				}
			case <-s.ctx.Done():
				return
			}
		}
	}()
}

func (s Sensor) Stop() {
	s.ctx.Done()
}

func (s Sensor) GetTemperatureAndHumidity() (float32, float32) {
	var temp, humid float32

	s.mtx.Lock()
	temp = *s.temp
	humid = *s.humid
	s.mtx.Unlock()

	return temp, humid
}

func (s Sensor) update() error {
	t, h, _, err := dht.ReadDHTxxWithContextAndRetry(s.ctx, dht.DHT22, sensorDataPin, false, sensorRetryLimit)
	if err != nil {
		return err
	}

	s.mtx.Lock()
	*s.temp = t
	*s.humid = h
	s.mtx.Unlock()

	s.tempGauge.Set(float64(t))
	s.humidGauge.Set(float64(h))

	return nil
}
