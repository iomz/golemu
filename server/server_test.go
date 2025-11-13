//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package server

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/iomz/go-llrp"
	"github.com/iomz/golemu/tag"
)

func TestNewServer(t *testing.T) {
	server := NewServer("127.0.0.1", 5084, 3000, 1500, 10000, 5, 1000, "test.gob")

	if server == nil {
		t.Fatal("NewServer returned nil")
	}
	if server.ip != "127.0.0.1" {
		t.Errorf("expected ip 127.0.0.1, got %s", server.ip)
	}
	if server.port != 5084 {
		t.Errorf("expected port 5084, got %d", server.port)
	}
	if server.apiPort != 3000 {
		t.Errorf("expected apiPort 3000, got %d", server.apiPort)
	}
	if server.pdu != 1500 {
		t.Errorf("expected pdu 1500, got %d", server.pdu)
	}
	if server.reportInterval != 10000 {
		t.Errorf("expected reportInterval 10000, got %d", server.reportInterval)
	}
	if server.keepaliveInterval != 5 {
		t.Errorf("expected keepaliveInterval 5, got %d", server.keepaliveInterval)
	}
	if server.initialMessageID != 1000 {
		t.Errorf("expected initialMessageID 1000, got %d", server.initialMessageID)
	}
	if server.file != "test.gob" {
		t.Errorf("expected file test.gob, got %s", server.file)
	}
	if server.tagManagerChan == nil {
		t.Error("tagManagerChan should not be nil")
	}
	if server.tagUpdatedChan == nil {
		t.Error("tagUpdatedChan should not be nil")
	}
	if server.tagService == nil {
		t.Error("tagService should not be nil")
	}
	if server.isConnAlive == nil {
		t.Error("isConnAlive should not be nil")
	}
	if server.llrpHandler == nil {
		t.Error("llrpHandler should not be nil")
	}
}

func TestServer_loadTags_FileNotExists(t *testing.T) {
	server := NewServer("127.0.0.1", 5084, 3000, 1500, 10000, 5, 1000, "nonexistent.gob")

	// Should not panic when file doesn't exist
	server.loadTags()

	tags := server.tagService.GetTags()
	if len(tags) != 0 {
		t.Errorf("expected 0 tags when file doesn't exist, got %d", len(tags))
	}
}

func TestServer_loadTags_InvalidFile(t *testing.T) {
	// Create a temporary file with invalid content
	tmpDir := t.TempDir()
	invalidFile := filepath.Join(tmpDir, "invalid.gob")
	err := os.WriteFile(invalidFile, []byte("invalid gob data"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	server := NewServer("127.0.0.1", 5084, 3000, 1500, 10000, 5, 1000, invalidFile)

	// Should not panic when file is invalid
	server.loadTags()

	tags := server.tagService.GetTags()
	if len(tags) != 0 {
		t.Errorf("expected 0 tags when file is invalid, got %d", len(tags))
	}
}

func TestServer_runTagManager(t *testing.T) {
	server := NewServer("127.0.0.1", 5084, 3000, 1500, 10000, 5, 1000, "test.gob")

	tag1, err := llrp.NewTag(&llrp.TagRecord{PCBits: "3000", EPC: "001100000111001000100111011000100111111100101110101001001000000000000000000000000001110001101010"})
	if err != nil {
		t.Fatalf("failed to create tag1: %v", err)
	}

	// Test AddTags command
	cmd := tag.Manager{
		Action: tag.AddTags,
		Tags:   llrp.Tags{tag1},
	}

	signals := make(chan os.Signal, 1)
	// Start runTagManager in a goroutine - it will run until a signal is received
	// We don't send a signal in tests to avoid log.Fatalf terminating the test process
	go server.runTagManager(signals)

	// Send command
	server.tagManagerChan <- cmd

	// Wait for response with timeout
	select {
	case result := <-server.tagManagerChan:
		if len(result.Tags) != 1 {
			t.Errorf("expected 1 tag in result, got %d", len(result.Tags))
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for tag manager response")
	}

	// Verify tag was added
	tags := server.tagService.GetTags()
	if len(tags) != 1 {
		t.Errorf("expected 1 tag in service, got %d", len(tags))
	}
	// Note: The goroutine will continue running, but that's fine for tests
}

func TestServer_runTagManager_RetrieveTags(t *testing.T) {
	server := NewServer("127.0.0.1", 5084, 3000, 1500, 10000, 5, 1000, "test.gob")

	tag1, err := llrp.NewTag(&llrp.TagRecord{PCBits: "3000", EPC: "001100000111001000100111011000100111111100101110101001001000000000000000000000000001110001101010"})
	if err != nil {
		t.Fatalf("failed to create tag1: %v", err)
	}

	// Set tags first
	server.tagService.SetTags(llrp.Tags{tag1})

	// Test RetrieveTags command
	cmd := tag.Manager{
		Action: tag.RetrieveTags,
		Tags:   llrp.Tags{},
	}

	signals := make(chan os.Signal, 1)
	// Start runTagManager in a goroutine - it will run until a signal is received
	// We don't send a signal in tests to avoid log.Fatalf terminating the test process
	go server.runTagManager(signals)

	// Send command
	server.tagManagerChan <- cmd

	// Wait for response with timeout
	select {
	case result := <-server.tagManagerChan:
		if len(result.Tags) != 1 {
			t.Errorf("expected 1 tag in result, got %d", len(result.Tags))
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for tag manager response")
	}
}

func TestServer_runTagManager_DeleteTags(t *testing.T) {
	server := NewServer("127.0.0.1", 5084, 3000, 1500, 10000, 5, 1000, "test.gob")

	tag1, err := llrp.NewTag(&llrp.TagRecord{PCBits: "3000", EPC: "001100000111001000100111011000100111111100101110101001001000000000000000000000000001110001101010"})
	if err != nil {
		t.Fatalf("failed to create tag1: %v", err)
	}

	// Set tags first
	server.tagService.SetTags(llrp.Tags{tag1})

	// Test DeleteTags command
	cmd := tag.Manager{
		Action: tag.DeleteTags,
		Tags:   llrp.Tags{tag1},
	}

	signals := make(chan os.Signal, 1)
	// Start runTagManager in a goroutine - it will run until a signal is received
	// We don't send a signal in tests to avoid log.Fatalf terminating the test process
	go server.runTagManager(signals)

	// Send command
	server.tagManagerChan <- cmd

	// Wait for response with timeout
	select {
	case result := <-server.tagManagerChan:
		if len(result.Tags) != 1 {
			t.Errorf("expected 1 tag in result, got %d", len(result.Tags))
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for tag manager response")
	}

	// Verify tag was deleted
	tags := server.tagService.GetTags()
	if len(tags) != 0 {
		t.Errorf("expected 0 tags after deletion, got %d", len(tags))
	}
}
