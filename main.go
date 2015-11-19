package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"
)

const currentVersion = "0.0.1"

var (
	flagset = flag.NewFlagSet("vulcand-healthcheck", flag.ExitOnError)
	flags   = struct {
		Version            bool
		Port               uint
		Path               string
		Timeout            time.Duration
		Interval           time.Duration
		UnhealthyThreshold uint
		HealthyThreshold   uint
		PrivateIP          string
		BackendID          string
		ServerID           string
	}{}
)

func usage() {
	fmt.Fprintf(os.Stderr, `
  NAME
  vulcand-healthcheck

  USAGE
  vulcand-healthcheck [options]

  OPTIONS
  `)
	flagset.PrintDefaults()
}

func init() {
	flagset.BoolVar(&flags.Version, "version", false, "Print the version and exit")
	flagset.BoolVar(&flags.Version, "v", false, "Print the version and exit")

	flagset.UintVar(&flags.Port, "port", 80, "Port exposed by the service")
	flagset.StringVar(&flags.Path, "path", "/", "Path hit to check the service healthiness (should return a 2XX status code)")
	flagset.DurationVar(&flags.Timeout, "timeout", 2*time.Second, "Time to wait when receiving a response from the health check")
	flagset.DurationVar(&flags.Interval, "interval", 5*time.Second, "Amount of time between health checks")
	flagset.UintVar(&flags.UnhealthyThreshold, "unhealthy-threshold", 2, "Number of consecutive health check failures before declaring the service unhealthy.")
	flagset.UintVar(&flags.HealthyThreshold, "healthy-threshold", 5, "Number of consecutive health check successes before declaring the service healthy.")
	flagset.StringVar(&flags.PrivateIP, "private-ip", "127.0.0.1", "IP address declared into vulcand")
	flagset.StringVar(&flags.BackendID, "backend-id", "", "Vulcand backend ID")
	flagset.StringVar(&flags.ServerID, "server-id", "", "Vulcand server ID")
}

func main() {
	flagset.Parse(os.Args[1:])
	flagset.Usage = usage

	if len(os.Args) < 2 {
		flagset.Usage()
		os.Exit(0)
	}

	if flags.Version {
		fmt.Printf("vulcand-healthcheck v%s", currentVersion)
		os.Exit(0)
	}

	if flags.Timeout > flags.Interval {
		fmt.Printf("The timeout duration cannot be greater than the interval")
		os.Exit(1)
	}

	vulcandURL := "http://localhost:8182"

	if v := os.Getenv("VULCAND_URL"); v != "" {
		vulcandURL = v
	}

	registry := NewRegistry(
		vulcandURL,
		flags.BackendID,
		flags.ServerID,
		flags.PrivateIP,
		flags.Port,
		flags.Interval+flags.Timeout,
	)

	healthCheck := NewHealthCheck(
		flags.PrivateIP,
		flags.Port,
		flags.Path,
		flags.Timeout,
	)

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Kill, os.Interrupt)

	stopChan := make(chan bool)

	watcher := NewWatcher(
		healthCheck,
		registry,
		flags.Interval,
		flags.HealthyThreshold,
		flags.UnhealthyThreshold,
		stopChan,
	)
	go watcher.Watch()

	<-sig
	log.Println("Termination signal received, will stop")

	stopChan <- true
}
