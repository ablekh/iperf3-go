package client

import (
	"encoding/json"
	"fmt"
	"iperf3-go/internal/protocol"
	"log"
	"net"
	"time"
)

// Config holds client configuration
type Config struct {
	Host      string
	Port      int
	Time      int
	Parallel  int
	Reverse   bool
	JSON      bool
	Verbose   bool
	Window    int
	Length    int
	Bandwidth int64
	Protocol  string
}

// Client represents an iperf3 client
type Client struct {
	config *Config
}

// New creates a new iperf3 client
func New(config *Config) *Client {
	return &Client{
		config: config,
	}
}

// Run starts the iperf3 client test
func (c *Client) Run() error {
	if c.config.Verbose {
		log.Printf("Connecting to host %s, port %d", c.config.Host, c.config.Port)
	}

	// Determine protocol type
	protocolType := c.config.Protocol
	if protocolType == "" {
		protocolType = "tcp"
	}

	// Connect to server
	addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)
	var conn net.Conn
	var err error

	switch protocolType {
	case "udp":
		conn, err = net.Dial("udp", addr)
	case "sctp":
		// SCTP support would require additional libraries
		return fmt.Errorf("SCTP protocol not yet implemented")
	default: // tcp
		conn, err = net.Dial("tcp", addr)
	}

	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", addr, err)
	}
	defer conn.Close()

	if c.config.Verbose {
		log.Printf("Connected to %s", conn.RemoteAddr())
	}

	// Send test configuration
	testConfig := &protocol.TestConfig{
		Protocol:  protocolType,
		Time:      c.config.Time,
		Parallel:  c.config.Parallel,
		Reverse:   c.config.Reverse,
		Window:    c.config.Window,
		Length:    c.config.Length,
		Bandwidth: c.config.Bandwidth,
	}

	configData, err := json.Marshal(testConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal test config: %w", err)
	}

	startMsg := &protocol.Message{
		Type: protocol.MessageTypeTestStart,
		Data: configData,
	}

	if err := protocol.WriteMessage(conn, startMsg); err != nil {
		return fmt.Errorf("failed to send test start: %w", err)
	}

	// Wait for acknowledgment
	ackMsg, err := protocol.ReadMessage(conn)
	if err != nil {
		return fmt.Errorf("failed to read test start ack: %w", err)
	}

	if ackMsg.Type != protocol.MessageTypeTestStartAck {
		return fmt.Errorf("unexpected message type: %d", ackMsg.Type)
	}

	if c.config.Verbose {
		log.Printf("Test started")
	}

	// Run the test
	return c.runTest(conn, protocolType)
}

// runTest runs the actual performance test
func (c *Client) runTest(conn net.Conn, protocolType string) error {
	duration := time.Duration(c.config.Time) * time.Second
	if duration == 0 {
		duration = 10 * time.Second // default
	}

	startTime := time.Now()
	var totalBytes int64
	buffer := make([]byte, 128*1024) // 128KB buffer

	// Fill buffer with test data
	for i := range buffer {
		buffer[i] = byte(i % 256)
	}

	if !c.config.JSON {
		if protocolType == "udp" {
			fmt.Printf("Connecting to host %s, port %d\n", c.config.Host, c.config.Port)
			fmt.Printf("[  4] local %s port %d connected to %s port %d\n",
				conn.LocalAddr().String(), getPort(conn.LocalAddr()),
				conn.RemoteAddr().String(), getPort(conn.RemoteAddr()))
			fmt.Printf("[ ID] Interval           Transfer     Bitrate         Jitter    Lost/Total Datagrams\n")
		} else {
			fmt.Printf("Connecting to host %s, port %d\n", c.config.Host, c.config.Port)
			fmt.Printf("[  4] local %s port %d connected to %s port %d\n",
				conn.LocalAddr().String(), getPort(conn.LocalAddr()),
				conn.RemoteAddr().String(), getPort(conn.RemoteAddr()))
			fmt.Printf("[ ID] Interval           Transfer     Bitrate\n")
		}
	}

	// Send data and collect interval results
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	intervalBytes := int64(0)
	intervalNum := 0

	// Start sending data in a goroutine
	go func() {
		if protocolType == "udp" {
			// For UDP, send packets at controlled rate
			packetSize := c.config.Length
			if packetSize == 0 {
				packetSize = 1470 // Default UDP payload size
			}
			udpBuffer := make([]byte, packetSize)
			copy(udpBuffer, buffer[:min(len(buffer), packetSize)])

			// Calculate target rate
			targetBandwidth := c.config.Bandwidth
			if targetBandwidth == 0 {
				targetBandwidth = 1000000 // 1 Mbps default for UDP
			}

			packetInterval := time.Duration(float64(packetSize*8) / float64(targetBandwidth) * float64(time.Second))
			ticker := time.NewTicker(packetInterval)
			defer ticker.Stop()

			for time.Since(startTime) < duration {
				select {
				case <-ticker.C:
					n, err := conn.Write(udpBuffer)
					if err != nil {
						return
					}
					totalBytes += int64(n)
					intervalBytes += int64(n)
				}
			}
		} else {
			// TCP - send as fast as possible
			for time.Since(startTime) < duration {
				n, err := conn.Write(buffer)
				if err != nil {
					return
				}
				totalBytes += int64(n)
				intervalBytes += int64(n)
			}
		}
	}()

	// Read interval messages from server
	for {
		select {
		case <-ticker.C:
			intervalNum++
			elapsed := time.Since(startTime).Seconds()

			if !c.config.JSON {
				transfer := float64(intervalBytes) / (1024 * 1024)  // MB
				bitrate := float64(intervalBytes*8) / (1024 * 1024) // Mbits/sec
				fmt.Printf("[  4] %7.2f-%7.2f sec  %7.2f MBytes  %7.2f Mbits/sec\n",
					elapsed-1, elapsed, transfer, bitrate)
			}

			intervalBytes = 0

			if elapsed >= duration.Seconds() {
				goto testComplete
			}

		case <-time.After(duration + 2*time.Second):
			goto testComplete
		}
	}

testComplete:
	elapsed := time.Since(startTime).Seconds()

	if c.config.JSON {
		// Output JSON results
		results := map[string]interface{}{
			"start": map[string]interface{}{
				"connected": []map[string]interface{}{
					{
						"socket":      4,
						"local_host":  conn.LocalAddr().String(),
						"local_port":  getPort(conn.LocalAddr()),
						"remote_host": conn.RemoteAddr().String(),
						"remote_port": getPort(conn.RemoteAddr()),
					},
				},
				"version":     "iperf3-go 1.0.0",
				"system_info": "Go implementation",
				"timestamp": map[string]interface{}{
					"time":     time.Now().Unix(),
					"timesecs": time.Now().Unix(),
				},
			},
			"end": map[string]interface{}{
				"streams": []map[string]interface{}{
					{
						"socket":          4,
						"start":           0,
						"end":             elapsed,
						"seconds":         elapsed,
						"bytes":           totalBytes,
						"bits_per_second": float64(totalBytes*8) / elapsed,
						"sender":          true,
					},
				},
				"sum_sent": map[string]interface{}{
					"start":           0,
					"end":             elapsed,
					"seconds":         elapsed,
					"bytes":           totalBytes,
					"bits_per_second": float64(totalBytes*8) / elapsed,
					"sender":          true,
				},
				"sum_received": map[string]interface{}{
					"start":           0,
					"end":             elapsed,
					"seconds":         elapsed,
					"bytes":           totalBytes,
					"bits_per_second": float64(totalBytes*8) / elapsed,
					"sender":          false,
				},
			},
		}

		jsonData, _ := json.MarshalIndent(results, "", "  ")
		fmt.Println(string(jsonData))
	} else {
		// Output standard format
		fmt.Printf("- - - - - - - - - - - - - - - - - - - - - - - - -\n")
		fmt.Printf("[ ID] Interval           Transfer     Bitrate\n")

		transfer := float64(totalBytes) / (1024 * 1024)            // MB
		bitrate := float64(totalBytes*8) / (1024 * 1024) / elapsed // Mbits/sec

		fmt.Printf("[  4] %7.2f-%7.2f sec  %7.2f MBytes  %7.2f Mbits/sec                  sender\n",
			0.0, elapsed, transfer, bitrate)
		fmt.Printf("[  4] %7.2f-%7.2f sec  %7.2f MBytes  %7.2f Mbits/sec                  receiver\n",
			0.0, elapsed, transfer, bitrate)
		fmt.Printf("\niperf Done.\n")
	}

	return nil
}

// Helper function to get port from address
func getPort(addr net.Addr) int {
	if tcpAddr, ok := addr.(*net.TCPAddr); ok {
		return tcpAddr.Port
	}
	if udpAddr, ok := addr.(*net.UDPAddr); ok {
		return udpAddr.Port
	}
	return 0
}

// Helper function to get minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
