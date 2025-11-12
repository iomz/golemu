//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/iomz/go-llrp"
	"github.com/iomz/golemu/tag"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestNewHandler(t *testing.T) {
	tagManagerChan := make(chan tag.Manager, 1)
	handler := NewHandler(tagManagerChan)

	if handler == nil {
		t.Fatal("NewHandler returned nil")
	}
	if handler.tagManagerChan != tagManagerChan {
		t.Error("tagManagerChan not set correctly")
	}
}

func TestHandler_PostTag_Success(t *testing.T) {
	tagManagerChan := make(chan tag.Manager, 10)
	handler := NewHandler(tagManagerChan)

	router := setupRouter()
	router.POST("/tags", handler.PostTag)

	// Create test tag data
	tagData := []llrp.TagRecord{
		{PCBits: "3000", EPC: "001100000111001000100111011000100111111100101110101001001000000000000000000000000001110001101010"},
	}
	jsonData, _ := json.Marshal(tagData)

	req, _ := http.NewRequest("POST", "/tags", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	// Check that message was sent to channel
	select {
	case <-tagManagerChan:
		// Expected
	default:
		t.Error("no message sent to tagManagerChan")
	}
}

func TestHandler_PostTag_InvalidJSON(t *testing.T) {
	tagManagerChan := make(chan tag.Manager, 1)
	handler := NewHandler(tagManagerChan)

	router := setupRouter()
	router.POST("/tags", handler.PostTag)

	req, _ := http.NewRequest("POST", "/tags", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandler_DeleteTag_Success(t *testing.T) {
	tagManagerChan := make(chan tag.Manager, 10)
	handler := NewHandler(tagManagerChan)

	router := setupRouter()
	router.DELETE("/tags", handler.DeleteTag)

	tagData := []llrp.TagRecord{
		{PCBits: "3000", EPC: "001100000111001000100111011000100111111100101110101001001000000000000000000000000001110001101010"},
	}
	jsonData, _ := json.Marshal(tagData)

	req, _ := http.NewRequest("DELETE", "/tags", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Check that message was sent to channel
	select {
	case <-tagManagerChan:
		// Expected
	default:
		t.Error("no message sent to tagManagerChan")
	}
}

func TestHandler_DeleteTag_InvalidJSON(t *testing.T) {
	tagManagerChan := make(chan tag.Manager, 1)
	handler := NewHandler(tagManagerChan)

	router := setupRouter()
	router.DELETE("/tags", handler.DeleteTag)

	req, _ := http.NewRequest("DELETE", "/tags", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandler_GetTags_Success(t *testing.T) {
	tagManagerChan := make(chan tag.Manager, 10)
	handler := NewHandler(tagManagerChan)

	// Set up a goroutine to handle the retrieve request
	// This simulates the tag manager service responding
	ready := make(chan bool)
	go func() {
		close(ready) // Signal that goroutine is ready
		cmd := <-tagManagerChan
		if cmd.Action == tag.RetrieveTags {
			tag1, _ := llrp.NewTag(&llrp.TagRecord{PCBits: "3000", EPC: "001100000111001000100111011000100111111100101110101001001000000000000000000000000001110001101010"})
			cmd.Tags = []*llrp.Tag{tag1}
			tagManagerChan <- cmd
		}
	}()

	// Wait for goroutine to be ready
	<-ready

	router := setupRouter()
	router.GET("/tags", handler.GetTags)

	req, _ := http.NewRequest("GET", "/tags", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("expected 1 tag, got %d", len(result))
	}
}

func TestHandler_reqAddTag(t *testing.T) {
	tagManagerChan := make(chan tag.Manager, 10)
	handler := NewHandler(tagManagerChan)

	tagData := []llrp.TagRecord{
		{PCBits: "3000", EPC: "001100000111001000100111011000100111111100101110101001001000000000000000000000000001110001101010"},
	}

	result := handler.reqAddTag(tagData)

	if result != "add" {
		t.Errorf("expected 'add', got '%s'", result)
	}

	// Check that message was sent
	select {
	case cmd := <-tagManagerChan:
		if cmd.Action != tag.AddTags {
			t.Errorf("expected AddTags action, got %v", cmd.Action)
		}
		if len(cmd.Tags) != 1 {
			t.Errorf("expected 1 tag, got %d", len(cmd.Tags))
		}
	default:
		t.Error("no message sent to tagManagerChan")
	}
}

func TestHandler_reqDeleteTag(t *testing.T) {
	tagManagerChan := make(chan tag.Manager, 10)
	handler := NewHandler(tagManagerChan)

	tagData := []llrp.TagRecord{
		{PCBits: "3000", EPC: "001100000111001000100111011000100111111100101110101001001000000000000000000000000001110001101010"},
	}

	result := handler.reqDeleteTag(tagData)

	if result != "delete" {
		t.Errorf("expected 'delete', got '%s'", result)
	}

	// Check that message was sent
	select {
	case cmd := <-tagManagerChan:
		if cmd.Action != tag.DeleteTags {
			t.Errorf("expected DeleteTags action, got %v", cmd.Action)
		}
		if len(cmd.Tags) != 1 {
			t.Errorf("expected 1 tag, got %d", len(cmd.Tags))
		}
	default:
		t.Error("no message sent to tagManagerChan")
	}
}

func TestHandler_reqRetrieveTag(t *testing.T) {
	tagManagerChan := make(chan tag.Manager, 10)
	handler := NewHandler(tagManagerChan)

	// Set up a goroutine to handle the retrieve request
	// This simulates the tag manager service responding
	ready := make(chan bool)
	go func() {
		close(ready) // Signal that goroutine is ready
		cmd := <-tagManagerChan
		if cmd.Action == tag.RetrieveTags {
			tag1, _ := llrp.NewTag(&llrp.TagRecord{PCBits: "3000", EPC: "001100000111001000100111011000100111111100101110101001001000000000000000000000000001110001101010"})
			tag2, _ := llrp.NewTag(&llrp.TagRecord{PCBits: "3000", EPC: "001100000111001000100111011000100111111100101110101001001000000000000000000000000001110001101011"})
			cmd.Tags = []*llrp.Tag{tag1, tag2}
			tagManagerChan <- cmd
		}
	}()

	// Wait for goroutine to be ready before calling reqRetrieveTag
	<-ready

	// reqRetrieveTag blocks until it receives a response from the channel
	result := handler.reqRetrieveTag()

	if len(result) != 2 {
		t.Errorf("expected 2 tags, got %d", len(result))
	}
}
