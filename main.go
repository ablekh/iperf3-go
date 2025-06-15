package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"iperf3-go/internal/client"
	"iperf3-go/internal/server"
)

func main() {
	var (
		// Common flags
		port    = flag.Int("p", 5201, "server port to listen on/connect to")
		verbose = flag.Bool("v", false, "verbose output")
		version = flag.Bool("version", false, "show version information and quit")

		// Client flags
		clientMode = flag.String("c", "", "run in client mode, connecting to <host>")
		time       = flag.Int("t", 10, "time in seconds to transmit for (default 10 secs)")
		parallel   = flag.Int("P", 1, "number of parallel client streams to run")
		reverse    = flag.Bool("R", false, "run in reverse mode (server sends, client receives)")
		jsonOutput = flag.Bool("J", false, "output in JSON format")
		window     = flag.Int("w", 0, "window size / socket buffer size")
		length     = flag.Int("l", 128*1024, "length of buffer to read or write (default 128 KB)")
		bandwidth  = flag.Int64("b", 0, "target bandwidth in bits/sec (0 for unlimited)")
		udp        = flag.Bool("u", false, "use UDP rather than TCP")
		sctp       = flag.Bool("sctp", false, "use SCTP rather than TCP")

		// Server flags
		bind   = flag.String("B", "", "bind to a specific interface")
		daemon = flag.Bool("D", false, "run the server as a daemon")
		oneOff = flag.Bool("1", false, "handle one client connection then exit")
	)
	flag.Parse()

	if *version {
		fmt.Println("iperf3-go 1.0.0")
		fmt.Println("Compatible with iperf 3.x")
		os.Exit(0)
	}

	// Check if running in client mode
	if *clientMode != "" {
		// Determine protocol
		protocol := "tcp"
		if *udp {
			protocol = "udp"
		} else if *sctp {
			protocol = "sctp"
		}

		// Client mode
		clientConfig := &client.Config{
			Host:      *clientMode,
			Port:      *port,
			Time:      *time,
			Parallel:  *parallel,
			Reverse:   *reverse,
			JSON:      *jsonOutput,
			Verbose:   *verbose,
			Window:    *window,
			Length:    *length,
			Bandwidth: *bandwidth,
			Protocol:  protocol,
		}

		c := client.New(clientConfig)
		if err := c.Run(); err != nil {
			log.Fatalf("Client failed: %v", err)
		}
	} else {
		// Server mode (default)
		// Determine protocol for server
		serverProtocol := "tcp"
		if *udp {
			serverProtocol = "udp"
		} else if *sctp {
			serverProtocol = "sctp"
		}

		serverConfig := &server.Config{
			Port:     *port,
			Bind:     *bind,
			Verbose:  *verbose,
			Daemon:   *daemon,
			OneOff:   *oneOff,
			Protocol: serverProtocol,
		}

		srv := server.New(serverConfig)
		if err := srv.Start(); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}
}
