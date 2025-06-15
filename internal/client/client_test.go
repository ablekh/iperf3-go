package client

import (
	"testing"
)

func TestClientConfig(t *testing.T) {
	config := &Config{
		Host:      "localhost",
		Port:      5201,
		Time:      10,
		Parallel:  1,
		Reverse:   false,
		JSON:      false,
		Verbose:   true,
		Window:    0,
		Length:    128 * 1024,
		Bandwidth: 0,
	}

	client := New(config)
	if client == nil {
		t.Fatal("New() returned nil")
	}

	if client.config.Host != "localhost" {
		t.Errorf("Expected host localhost, got %s", client.config.Host)
	}

	if client.config.Port != 5201 {
		t.Errorf("Expected port 5201, got %d", client.config.Port)
	}

	if client.config.Time != 10 {
		t.Errorf("Expected time 10, got %d", client.config.Time)
	}

	if !client.config.Verbose {
		t.Error("Expected verbose to be true")
	}
}

func TestGetPort(t *testing.T) {
	// Test with nil address
	port := getPort(nil)
	if port != 0 {
		t.Errorf("Expected port 0 for nil address, got %d", port)
	}
}
