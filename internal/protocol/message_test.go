package protocol

import (
	"bytes"
	"net"
	"testing"
	"time"
)

func TestMessageReadWrite(t *testing.T) {
	// Create a pipe for testing
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	// Test message
	originalMsg := &Message{
		Type: MessageTypeTestStart,
		Data: []byte(`{"protocol":"tcp","time":10}`),
	}

	// Write message in a goroutine
	go func() {
		if err := WriteMessage(client, originalMsg); err != nil {
			t.Errorf("WriteMessage failed: %v", err)
		}
		client.Close()
	}()

	// Read message
	readMsg, err := ReadMessage(server)
	if err != nil {
		t.Fatalf("ReadMessage failed: %v", err)
	}

	// Verify message
	if readMsg.Type != originalMsg.Type {
		t.Errorf("Message type mismatch: got %d, want %d", readMsg.Type, originalMsg.Type)
	}

	if !bytes.Equal(readMsg.Data, originalMsg.Data) {
		t.Errorf("Message data mismatch: got %s, want %s", readMsg.Data, originalMsg.Data)
	}
}

func TestMessageReadTimeout(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	// Set read timeout
	server.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

	// Try to read without writing anything
	_, err := ReadMessage(server)
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
}

func TestMessageTooLarge(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	// Create a message that's too large
	largeData := make([]byte, 2*1024*1024) // 2MB
	largeMsg := &Message{
		Type: MessageTypeTestStart,
		Data: largeData,
	}

	// Write large message in a goroutine
	go func() {
		WriteMessage(client, largeMsg)
		client.Close()
	}()

	// Try to read large message
	_, err := ReadMessage(server)
	if err == nil {
		t.Error("Expected error for large message, got nil")
	}
}
