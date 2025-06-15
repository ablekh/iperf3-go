package server

import (
	"testing"
	"time"
)

func TestServerConfig(t *testing.T) {
	config := &Config{
		Port:    8080,
		Bind:    "localhost",
		Verbose: true,
		OneOff:  true,
	}

	server := New(config)
	if server == nil {
		t.Fatal("New() returned nil")
	}

	if server.config.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", server.config.Port)
	}

	if server.config.Bind != "localhost" {
		t.Errorf("Expected bind localhost, got %s", server.config.Bind)
	}

	if !server.config.Verbose {
		t.Error("Expected verbose to be true")
	}

	if !server.config.OneOff {
		t.Error("Expected oneOff to be true")
	}
}

func TestGenerateSessionID(t *testing.T) {
	id1 := generateSessionID()
	time.Sleep(1 * time.Millisecond) // Ensure different timestamps
	id2 := generateSessionID()

	if id1 == id2 {
		t.Error("Expected different session IDs")
	}

	if len(id1) == 0 {
		t.Error("Expected non-empty session ID")
	}
}

func TestGetPort(t *testing.T) {
	// This test would require creating actual network addresses
	// For now, we'll test the basic functionality
	port := getPort(nil)
	if port != 0 {
		t.Errorf("Expected port 0 for nil address, got %d", port)
	}
}
