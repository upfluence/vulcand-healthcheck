package main

import (
	"log"
	"time"
)

type Watcher struct {
	registry           *Registry
	healthCheck        *HealthCheck
	interval           time.Duration
	stopChan           chan bool
	healthyThreshold   uint
	unhealthyThreshold uint

	history      []Status
	isRegistered bool
}

func NewWatcher(
	healthCheck *HealthCheck,
	registry *Registry,
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
		[]Status{},
		false,
	}
}

func (w *Watcher) Watch() {
	for {
		select {
		case <-time.After(w.interval):
			log.Println("Check if the service is healthy")

			r := w.healthCheck.Ping()

			if r == Healthy {
				log.Println("The service is healthy")
			} else {
				log.Println("The service is not healthy")
			}

			w.history = append(w.history, r)

			if len(w.history) > int(w.unhealthyThreshold) &&
				len(w.history) > int(w.healthyThreshold) {
				w.history = w.history[1:]
			}

			if w.shouldDeleteServer() {
				err := w.registry.DeleteServer()

				if err != nil {
					log.Printf("DeleteServer: %s", err.Error())
				}
			} else if w.shouldRegisterServer() {
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
		if status == Unhealthy {
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
		if status == Healthy {
			return false
		}
	}

	return true
}
