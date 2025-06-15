package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

// Message types for iperf3 protocol
const (
	MessageTypeTestStart    = 1
	MessageTypeTestStartAck = 2
	MessageTypeTestRunning  = 3
	MessageTypeTestEnd      = 4
	MessageTypeInterval     = 5
	MessageTypeError        = 6
)

// Message represents an iperf3 protocol message
type Message struct {
	Type int
	Data []byte
}

// ReadMessage reads an iperf3 message from a connection
func ReadMessage(conn net.Conn) (*Message, error) {
	// Read message length (4 bytes, big endian)
	var length uint32
	if err := binary.Read(conn, binary.BigEndian, &length); err != nil {
		return nil, fmt.Errorf("failed to read message length: %w", err)
	}

	if length > 1024*1024 { // 1MB limit
		return nil, fmt.Errorf("message too large: %d bytes", length)
	}

	// Read message type (4 bytes, big endian)
	var msgType uint32
	if err := binary.Read(conn, binary.BigEndian, &msgType); err != nil {
		return nil, fmt.Errorf("failed to read message type: %w", err)
	}

	// Read message data
	data := make([]byte, length-4) // subtract 4 for the type field
	if _, err := io.ReadFull(conn, data); err != nil {
		return nil, fmt.Errorf("failed to read message data: %w", err)
	}

	return &Message{
		Type: int(msgType),
		Data: data,
	}, nil
}

// WriteMessage writes an iperf3 message to a connection
func WriteMessage(conn net.Conn, msg *Message) error {
	// Calculate total length (4 bytes for type + data length)
	totalLength := uint32(4 + len(msg.Data))

	// Write message length
	if err := binary.Write(conn, binary.BigEndian, totalLength); err != nil {
		return fmt.Errorf("failed to write message length: %w", err)
	}

	// Write message type
	if err := binary.Write(conn, binary.BigEndian, uint32(msg.Type)); err != nil {
		return fmt.Errorf("failed to write message type: %w", err)
	}

	// Write message data
	if _, err := conn.Write(msg.Data); err != nil {
		return fmt.Errorf("failed to write message data: %w", err)
	}

	return nil
}