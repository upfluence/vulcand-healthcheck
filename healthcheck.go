package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
	Healthy = iota
	Unhealthy
)

type Status int

type HealthCheck struct {
	url     string
	timeout time.Duration
}

func NewHealthCheck(ipAddr string, port uint, path string, timeout time.Duration) *HealthCheck {
	url := fmt.Sprintf("http://%s:%d", ipAddr, port)

	if len(path) > 0 && path[0:1] == "/" {
		url += path
	} else {
		url += "/" + path
	}

	return &HealthCheck{url, timeout}
}

func (h *HealthCheck) Ping() Status {
	responseChan := make(chan Status)

	go func() {
		responseChan <- h.ping()
	}()

	select {
	case <-time.After(h.timeout):
		return Unhealthy
	case s := <-responseChan:
		return s
	}
}

func (h *HealthCheck) ping() Status {
	log.Printf("Fetching: %s", h.url)
	// TODO: Be able to choose the HTTP method
	r, err := http.Get(h.url)

	if err != nil {
		log.Printf(err.Error())
		return Unhealthy
	}

	// TODO: Maybe accept other status codes
	if r.StatusCode/100 == 2 {
		return Healthy
	}

	return Unhealthy
}
