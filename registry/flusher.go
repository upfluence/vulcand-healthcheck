package registry

import "net/http"

type ConnFlusher struct {
	Next *Registry
}

func (f *ConnFlusher) RegisterServer() error {
	return flushConns(f.Next.RegisterServer())
}

func (f *ConnFlusher) DeleteServer() error {
	return flushConns(f.Next.RegisterServer())
}

func flushConns(err error) error {
	if err != nil {
		http.DefaultClient.CloseIdleConnections()
	}

	return err
}
