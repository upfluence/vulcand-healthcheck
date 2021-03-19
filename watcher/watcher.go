package watcher

import (
	"log"
	"time"

	"github.com/upfluence/vulcand-healthcheck/healthcheck"
)

type Registry interface {
	RegisterServer() error
	DeleteServer() error
}

type Watcher struct {
	registry           Registry
	healthCheck        *healthcheck.HealthCheck
	interval           time.Duration
	stopChan           chan bool
	healthyThreshold   uint
	unhealthyThreshold uint

	history      []healthcheck.Status
	isRegistered bool
}

func NewWatcher(
	healthCheck *healthcheck.HealthCheck,
	registry Registry,
	interval time.Duration,
	healthyThreshold, unhealthyThreshold uint,
	stopChan chan bool) *Watcher {
	return &Watcher{
		registry,
		healthCheck,
		interval,
		stopChan,
		healthyThreshold,
		unhealthyThreshold,
		[]healthcheck.Status{},
		false,
	}
}

func (w *Watcher) Watch() {
	for {
		select {
		case <-time.After(w.interval):
			log.Println("Check if the service is healthy")

			r := w.healthCheck.Ping()

			if r == healthcheck.Healthy {
				log.Println("The service is healthy")
			} else {
				log.Println("The service is not healthy")
			}

			w.history = append(w.history, r)

			if len(w.history) > int(w.unhealthyThreshold) &&
				len(w.history) > int(w.healthyThreshold) {
				w.history = w.history[1:]
			}

			log.Printf("Health history: %+v", w.history)

			if w.shouldDeleteServer() {
				log.Println("Will unregister the server")
				err := w.registry.DeleteServer()

				if err != nil {
					log.Printf("DeleteServer: %s", err.Error())
				}
			} else if w.shouldRegisterServer() {
				log.Println("Will register the server")
				err := w.registry.RegisterServer()

				if err != nil {
					log.Printf("RegisterServer: %s", err.Error())
				}
			}
		case <-w.stopChan:
			log.Println("Gracefull stop asked")
			err := w.registry.DeleteServer()

			if err != nil {
				log.Printf("DeleteServer: %s", err.Error())
			}
		}
	}
}

func (w *Watcher) shouldRegisterServer() bool {
	if len(w.history) < int(w.healthyThreshold) {
		return false
	}

	for _, status := range w.history {
		if status == healthcheck.Unhealthy {
			return false
		}
	}

	return true
}

func (w *Watcher) shouldDeleteServer() bool {
	if len(w.history) < int(w.unhealthyThreshold) {
		return false
	}

	for _, status := range w.history {
		if status == healthcheck.Healthy {
			return false
		}
	}

	return true
}
