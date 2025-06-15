package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"iperf3-go/internal/server"
)

func main() {
	var (
		port     = flag.Int("p", 5201, "server port to listen on")
		bind     = flag.String("B", "", "bind to a specific interface")
		verbose  = flag.Bool("v", false, "verbose output")
		daemon   = flag.Bool("D", false, "run the server as a daemon")
		oneOff   = flag.Bool("1", false, "handle one client connection then exit")
		version  = flag.Bool("version", false, "show version information and quit")
	)
	flag.Parse()

	if *version {
		fmt.Println("iperf3-go 1.0.0")
		fmt.Println("Compatible with iperf 3.x")
		os.Exit(0)
	}

	config := &server.Config{
		Port:    *port,
		Bind:    *bind,
		Verbose: *verbose,
		Daemon:  *daemon,
		OneOff:  *oneOff,
	}

	srv := server.New(config)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}