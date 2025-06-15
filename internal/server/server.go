package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"iperf3-go/internal/protocol"
)

// Config holds server configuration
type Config struct {
	Port    int
	Bind    string
	Verbose bool
	Daemon  bool
	OneOff  bool
}

// Server represents an iperf3 server
type Server struct {
	config   *Config
	listener net.Listener
	sessions map[string]*Session
	mutex    sync.RWMutex
}

// Session represents a client test session
type Session struct {
	ID       string
	Conn     net.Conn
	Config   *protocol.TestConfig
	Results  *protocol.TestResults
	StartTime time.Time
}

// New creates a new iperf3 server
func New(config *Config) *Server {
	return &Server{
		config:   config,
		sessions: make(map[string]*Session),
	}
}

// Start starts the iperf3 server
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.config.Port)
	if s.config.Bind != "" {
		addr = fmt.Sprintf("%s:%d", s.config.Bind, s.config.Port)
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	s.listener = listener

	if s.config.Verbose {
		log.Printf("Server listening on %s", addr)
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
				Socket: 1,
				Start:  elapsed - 1,
				End:    elapsed,
				Seconds: 1.0,
				Bytes:  intervalBytes,
				BitsPerSecond: float64(intervalBytes * 8),
				Omitted: false,
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
				Socket: 1,
				Start:  0,
				End:    elapsed,
				Seconds: elapsed,
				Bytes:  totalBytes,
				BitsPerSecond: float64(totalBytes * 8) / elapsed,
				Sender: false,
			},
		},
		SumSent: protocol.StreamResult{
			Start:   0,
			End:     elapsed,
			Seconds: elapsed,
			Bytes:   totalBytes,
			BitsPerSecond: float64(totalBytes * 8) / elapsed,
			Sender: false,
		},
		SumReceived: protocol.StreamResult{
			Start:   0,
			End:     elapsed,
			Seconds: elapsed,
			Bytes:   totalBytes,
			BitsPerSecond: float64(totalBytes * 8) / elapsed,
			Sender: false,
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