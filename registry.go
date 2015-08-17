package main

import (
	"fmt"
	"time"

	"github.com/mailgun/vulcand/api"
	"github.com/mailgun/vulcand/engine"
	"github.com/mailgun/vulcand/plugin/registry"
)

type Registry struct {
	client    *api.Client
	backendID string
	serverID  string
	url       string
	ttl       time.Duration
}

func NewRegistry(
	addr, backendID, serverID, ipAddress string, port uint,
	ttl time.Duration) *Registry {
	return &Registry{
		api.NewClient(addr, registry.GetRegistry()),
		backendID,
		serverID,
		fmt.Sprintf("http://%s:%d", ipAddress, port),
		ttl,
	}
}

func (r *Registry) RegisterServer() error {
	s, err := engine.NewServer(r.serverID, r.url)

	if err != nil {
		return err
	}

	return r.client.UpsertServer(
		engine.BackendKey{Id: r.backendID},
		*s,
		r.ttl,
	)
}
func (r *Registry) DeleteServer() error {
	sk := engine.ServerKey{
		BackendKey: engine.BackendKey{Id: r.backendID},
		Id:         r.serverID,
	}

	return r.client.DeleteServer(sk)
}
