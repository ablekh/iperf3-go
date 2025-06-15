package server

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"iperf3-go/internal/protocol"

	"github.com/ishidawataru/sctp"
)

// Config holds server configuration
type Config struct {
	Port     int
	Bind     string
	Verbose  bool
	Daemon   bool
	OneOff   bool
	Protocol string
}

// Server represents an iperf3 server
type Server struct {
	config      *Config
	listener    net.Listener
	sessions    map[string]*Session
	udpSessions map[string]*protocol.UDPStats
	mutex       sync.RWMutex
}

// Session represents a client test session
type Session struct {
	ID        string
	Conn      net.Conn
	Config    *protocol.TestConfig
	Results   *protocol.TestResults
	StartTime time.Time
	UDPStats  *protocol.UDPStats
}

// New creates a new iperf3 server
func New(config *Config) *Server {
	return &Server{
		config:      config,
		sessions:    make(map[string]*Session),
		udpSessions: make(map[string]*protocol.UDPStats),
	}
}

// Start starts the iperf3 server
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.config.Port)
	if s.config.Bind != "" {
		addr = fmt.Sprintf("%s:%d", s.config.Bind, s.config.Port)
	}

	protocol := s.config.Protocol
	if protocol == "" {
		protocol = "tcp"
	}

	switch protocol {
	case "udp":
		return s.startUDPServer(addr)
	case "sctp":
		return s.startSCTPServer(addr)
	default: // tcp
		return s.startTCPServer(addr)
	}
}

// startTCPServer starts a TCP server
func (s *Server) startTCPServer(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	s.listener = listener

	if s.config.Verbose {
		log.Printf("TCP Server listening on %s", addr)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		go s.handleConnection(conn)

		if s.config.OneOff {
			break
		}
	}

	return nil
}

// startUDPServer starts a UDP server
func (s *Server) startUDPServer(addr string) error {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address %s: %w", addr, err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on UDP %s: %w", addr, err)
	}
	defer conn.Close()

	if s.config.Verbose {
		log.Printf("UDP Server listening on %s", addr)
	}

	// For UDP, we handle packets differently
	buffer := make([]byte, 65536) // Max UDP packet size
	for {
		n, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("Failed to read UDP packet: %v", err)
			continue
		}

		// Handle UDP packet in a goroutine
		go s.handleUDPPacket(conn, clientAddr, buffer[:n])

		if s.config.OneOff {
			break
		}
	}

	return nil
}

// startSCTPServer starts an SCTP server
func (s *Server) startSCTPServer(addr string) error {
	sctpAddr, err := sctp.ResolveSCTPAddr("sctp", addr)
	if err != nil {
		return fmt.Errorf("failed to resolve SCTP address %s: %w", addr, err)
	}

	listener, err := sctp.ListenSCTP("sctp", sctpAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on SCTP %s: %w", addr, err)
	}
	defer listener.Close()

	if s.config.Verbose {
		log.Printf("SCTP Server listening on %s", addr)
	}

	for {
		conn, err := listener.AcceptSCTP()
		if err != nil {
			log.Printf("Failed to accept SCTP connection: %v", err)
			continue
		}

		go s.handleConnection(conn)

		if s.config.OneOff {
			break
		}
	}

	return nil
}

// handleUDPPacket handles a UDP packet from a client with advanced statistics
func (s *Server) handleUDPPacket(conn *net.UDPConn, clientAddr *net.UDPAddr, data []byte) {
	if s.config.Verbose {
		log.Printf("Received UDP packet from %s, size: %d bytes", clientAddr, len(data))
	}

	clientKey := clientAddr.String()

	// Parse UDP packet header if present
	if len(data) >= 16 {
		// Extract packet header
		sequence := binary.BigEndian.Uint32(data[0:4])
		timestamp := int64(binary.BigEndian.Uint64(data[4:12]))
		magic := binary.BigEndian.Uint32(data[12:16])

		// Verify magic number
		const expectedMagic uint32 = 0x12345678
		if magic == expectedMagic {
			// Calculate statistics
			arrivalTime := time.Now().UnixNano()
			transitTime := float64(arrivalTime-timestamp) / 1000000.0 // Convert to milliseconds

			// Get or create UDP session stats
			s.mutex.Lock()
			stats, exists := s.udpSessions[clientKey]
			if !exists {
				stats = &protocol.UDPStats{
					LastSequence: sequence - 1, // Initialize to expect this sequence
				}
				s.udpSessions[clientKey] = stats
			}

			// Update packet statistics
			stats.TotalPackets++

			// Check for packet loss (sequence gaps)
			expectedSeq := stats.LastSequence + 1
			if sequence > expectedSeq {
				// Packets were lost
				lostCount := int64(sequence - expectedSeq)
				stats.LostPackets += lostCount
				if s.config.Verbose {
					log.Printf("Packet loss detected: expected seq %d, got %d (lost %d packets)",
						expectedSeq, sequence, lostCount)
				}
			} else if sequence < expectedSeq {
				// Out-of-order packet
				stats.OutOfOrder++
				if s.config.Verbose {
					log.Printf("Out-of-order packet: expected seq %d, got %d", expectedSeq, sequence)
				}
			}

			// Update sequence tracking
			if sequence >= stats.LastSequence {
				stats.LastSequence = sequence
			}

			// Calculate jitter using RFC 1889 algorithm
			if stats.LastArrivalTime > 0 {
				// Calculate transit time difference
				transitDiff := transitTime - stats.LastTransitTime
				if transitDiff < 0 {
					transitDiff = -transitDiff
				}

				// Update jitter using exponential smoothing (RFC 1889)
				// J(i) = J(i-1) + (|D(i-1,i)| - J(i-1))/16
				if stats.JitterCount == 0 {
					stats.JitterSum = transitDiff
				} else {
					currentJitter := stats.JitterSum / float64(stats.JitterCount)
					newJitter := currentJitter + (transitDiff-currentJitter)/16.0
					stats.JitterSum = newJitter * float64(stats.JitterCount+1)
				}
				stats.JitterCount++
			}

			// Update timing for next jitter calculation
			stats.LastArrivalTime = arrivalTime
			stats.LastTransitTime = transitTime

			s.mutex.Unlock()

			if s.config.Verbose {
				lossPercent := 0.0
				if stats.TotalPackets > 0 {
					lossPercent = float64(stats.LostPackets) / float64(stats.TotalPackets+stats.LostPackets) * 100.0
				}
				jitter := 0.0
				if stats.JitterCount > 0 {
					jitter = stats.JitterSum / float64(stats.JitterCount)
				}
				log.Printf("UDP stats from %s: seq=%d, transit=%.2fms, loss=%.2f%%, jitter=%.2fms, ooo=%d",
					clientAddr, sequence, transitTime, lossPercent, jitter, stats.OutOfOrder)
			}

			// Send acknowledgment with basic stats
			response := fmt.Sprintf("UDP packet received: seq=%d, total=%d, lost=%d, ooo=%d",
				sequence, stats.TotalPackets, stats.LostPackets, stats.OutOfOrder)
			_, err := conn.WriteToUDP([]byte(response), clientAddr)
			if err != nil {
				log.Printf("Failed to send UDP response to %s: %v", clientAddr, err)
			}
			return
		}
	}

	// Handle packets without proper header (legacy mode)
	response := []byte("UDP packet received (legacy mode)")
	_, err := conn.WriteToUDP(response, clientAddr)
	if err != nil {
		log.Printf("Failed to send UDP response to %s: %v", clientAddr, err)
	}
}

// handleConnection handles a client connection
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	if s.config.Verbose {
		log.Printf("New connection from %s", conn.RemoteAddr())
	}

	// Create new session
	session := &Session{
		ID:        generateSessionID(),
		Conn:      conn,
		StartTime: time.Now(),
	}

	s.mutex.Lock()
	s.sessions[session.ID] = session
	s.mutex.Unlock()

	defer func() {
		s.mutex.Lock()
		delete(s.sessions, session.ID)
		s.mutex.Unlock()
	}()

	// Handle iperf3 protocol
	if err := s.handleProtocol(session); err != nil {
		log.Printf("Protocol error for session %s: %v", session.ID, err)
	}
}

// handleProtocol handles the iperf3 protocol exchange
func (s *Server) handleProtocol(session *Session) error {
	// Read initial message from client
	msg, err := protocol.ReadMessage(session.Conn)
	if err != nil {
		return fmt.Errorf("failed to read initial message: %w", err)
	}

	switch msg.Type {
	case protocol.MessageTypeTestStart:
		return s.handleTestStart(session, msg)
	default:
		return fmt.Errorf("unexpected message type: %d", msg.Type)
	}
}

// handleTestStart handles test start message
func (s *Server) handleTestStart(session *Session, msg *protocol.Message) error {
	// Parse test configuration
	var config protocol.TestConfig
	if err := json.Unmarshal(msg.Data, &config); err != nil {
		return fmt.Errorf("failed to parse test config: %w", err)
	}

	session.Config = &config

	if s.config.Verbose {
		log.Printf("Test config: %+v", config)
	}

	// Send acknowledgment
	ack := &protocol.Message{
		Type: protocol.MessageTypeTestStartAck,
		Data: []byte("{}"),
	}

	if err := protocol.WriteMessage(session.Conn, ack); err != nil {
		return fmt.Errorf("failed to send test start ack: %w", err)
	}

	// Run the test
	return s.runTest(session)
}

// runTest runs the actual performance test
func (s *Server) runTest(session *Session) error {
	switch session.Config.Protocol {
	case "tcp", "":
		return s.runTCPTest(session)
	case "udp":
		return s.runUDPTest(session)
	default:
		return fmt.Errorf("unsupported protocol: %s", session.Config.Protocol)
	}
}

// runTCPTest runs a TCP performance test
func (s *Server) runTCPTest(session *Session) error {
	results := &protocol.TestResults{
		Start: protocol.TestStart{
			Connected: []protocol.Connection{
				{
					Socket:     1,
					LocalHost:  session.Conn.LocalAddr().String(),
					LocalPort:  s.config.Port,
					RemoteHost: session.Conn.RemoteAddr().String(),
					RemotePort: getPort(session.Conn.RemoteAddr()),
				},
			},
			Version:    "iperf3-go 1.0.0",
			SystemInfo: "Go implementation",
			Timestamp: protocol.Timestamp{
				Time:     time.Now().Unix(),
				Timesecs: time.Now().Unix(),
			},
			ConnectingTo: protocol.ConnectingTo{
				Host: session.Conn.RemoteAddr().String(),
				Port: getPort(session.Conn.RemoteAddr()),
			},
			Cookie: session.ID,
		},
	}

	session.Results = results

	// Send test results periodically during the test
	duration := time.Duration(session.Config.Time) * time.Second
	if duration == 0 {
		duration = 10 * time.Second // default
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	startTime := time.Now()
	var totalBytes int64
	intervalBytes := int64(0)
	intervalNum := 0

	buffer := make([]byte, 128*1024) // 128KB buffer

	go func() {
		for {
			n, err := session.Conn.Read(buffer)
			if err != nil {
				return
			}
			totalBytes += int64(n)
			intervalBytes += int64(n)
		}
	}()

	for {
		select {
		case <-ticker.C:
			intervalNum++
			elapsed := time.Since(startTime).Seconds()

			interval := protocol.Interval{
				Socket:        1,
				Start:         elapsed - 1,
				End:           elapsed,
				Seconds:       1.0,
				Bytes:         intervalBytes,
				BitsPerSecond: float64(intervalBytes * 8),
				Omitted:       false,
			}

			intervalMsg := &protocol.Message{
				Type: protocol.MessageTypeInterval,
				Data: mustMarshal(interval),
			}

			if err := protocol.WriteMessage(session.Conn, intervalMsg); err != nil {
				return fmt.Errorf("failed to send interval: %w", err)
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
	// Send final results
	elapsed := time.Since(startTime).Seconds()

	results.End = protocol.TestEnd{
		Streams: []protocol.StreamResult{
			{
				Socket:        1,
				Start:         0,
				End:           elapsed,
				Seconds:       elapsed,
				Bytes:         totalBytes,
				BitsPerSecond: float64(totalBytes*8) / elapsed,
				Sender:        false,
			},
		},
		SumSent: protocol.StreamResult{
			Start:         0,
			End:           elapsed,
			Seconds:       elapsed,
			Bytes:         totalBytes,
			BitsPerSecond: float64(totalBytes*8) / elapsed,
			Sender:        false,
		},
		SumReceived: protocol.StreamResult{
			Start:         0,
			End:           elapsed,
			Seconds:       elapsed,
			Bytes:         totalBytes,
			BitsPerSecond: float64(totalBytes*8) / elapsed,
			Sender:        false,
		},
		CPUUtilizationPercent: protocol.CPUUtilization{
			HostTotal:    0.0,
			HostUser:     0.0,
			HostSystem:   0.0,
			RemoteTotal:  0.0,
			RemoteUser:   0.0,
			RemoteSystem: 0.0,
		},
	}

	endMsg := &protocol.Message{
		Type: protocol.MessageTypeTestEnd,
		Data: mustMarshal(results.End),
	}

	return protocol.WriteMessage(session.Conn, endMsg)
}

// runUDPTest runs a UDP performance test
func (s *Server) runUDPTest(session *Session) error {
	// UDP test implementation would go here
	// For now, return an error as it's not implemented
	return fmt.Errorf("UDP tests not yet implemented")
}

// GetUDPStats returns UDP statistics for a client
func (s *Server) GetUDPStats(clientAddr string) *protocol.UDPStats {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if stats, exists := s.udpSessions[clientAddr]; exists {
		// Return a copy to avoid race conditions
		statsCopy := *stats
		return &statsCopy
	}
	return nil
}

// CleanupUDPSession removes UDP session statistics for a client
func (s *Server) CleanupUDPSession(clientAddr string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.udpSessions, clientAddr)
}

// Helper functions
func generateSessionID() string {
	return fmt.Sprintf("iperf3-go-%d", time.Now().UnixNano())
}

func getPort(addr net.Addr) int {
	if tcpAddr, ok := addr.(*net.TCPAddr); ok {
		return tcpAddr.Port
	}
	return 0
}

func mustMarshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}
