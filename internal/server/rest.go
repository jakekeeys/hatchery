package server

import (
	"context"
	"encoding/json"
	"github.com/gorilla/handlers"
	"github.com/jakekeeys/hatchery/internal/relay"
	"github.com/jakekeeys/hatchery/internal/sensor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
)

type Server struct {
	httpSrv *http.Server
	sensor  *sensor.Sensor
	relay   *relay.Relay
}

func New(sensor *sensor.Sensor, relay *relay.Relay) *Server {
	var s Server

	s.sensor = sensor
	s.relay = relay

	r := http.NewServeMux()
	r.Handle("/state", http.HandlerFunc(s.stateHandler))
	r.Handle("/relay/on", http.HandlerFunc(s.relayOnHandler))
	r.Handle("/relay/off", http.HandlerFunc(s.relayOffHandler))
	r.Handle("/metrics", promhttp.Handler())

	server := http.Server{
		Addr:    ":8080",
		Handler: handlers.LoggingHandler(os.Stdout, r),
	}

	s.httpSrv = &server

	return &s
}

func (s Server) Start() {
	go func() {
		err := s.httpSrv.ListenAndServe()
		if err != nil {
			logrus.Panic(err)
		}
	}()
}

func (s Server) Stop() {
	err := s.httpSrv.Shutdown(context.Background())
	if err != nil {
		logrus.Warn(err)
	}
}

type stateResponse struct {
	Temperature float32
	Humidity    float32
	RelayOn		bool
}

func (s Server) stateHandler(w http.ResponseWriter, r *http.Request) {
	t, h := s.sensor.GetTemperatureAndHumidity()
	on := s.relay.IsOn()

	resp := stateResponse{
		Temperature: t,
		Humidity:    h,
		RelayOn: on,
	}

	bytes, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	_, err = w.Write(bytes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s Server) relayOnHandler(w http.ResponseWriter, r *http.Request) {
	s.relay.SetOn()
}

func (s Server) relayOffHandler(w http.ResponseWriter, r *http.Request) {
	s.relay.SetOff()
}
