//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package api

import (
	"bytes"
	"encoding/json"
	"errors"
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

	// Create a tag to simulate successful addition
	tag1, err := llrp.NewTag(&llrp.TagRecord{PCBits: "3000", EPC: "001100000111001000100111011000100111111100101110101001001000000000000000000000000001110001101010"})
	if err != nil {
		t.Fatalf("failed to create tag1: %v", err)
	}

	// Set up a goroutine to handle the add request and respond
	ready := make(chan bool)
	go func() {
		close(ready)
		cmd := <-tagManagerChan
		if cmd.Action == tag.AddTags {
			// Simulate successful addition by returning the tag
			cmd.Tags = []*llrp.Tag{tag1}
			tagManagerChan <- cmd
		}
	}()

	// Wait for goroutine to be ready
	<-ready

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

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response["message"] == nil {
		t.Error("expected message field in response")
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

	// Create a tag to simulate it being in storage
	tag1, err := llrp.NewTag(&llrp.TagRecord{PCBits: "3000", EPC: "001100000111001000100111011000100111111100101110101001001000000000000000000000000001110001101010"})
	if err != nil {
		t.Fatalf("failed to create tag1: %v", err)
	}

	// Set up a goroutine to handle the delete request and respond
	ready := make(chan bool)
	go func() {
		close(ready)
		cmd := <-tagManagerChan
		if cmd.Action == tag.DeleteTags {
			// Simulate successful deletion by returning the tag
			cmd.Tags = []*llrp.Tag{tag1}
			tagManagerChan <- cmd
		}
	}()

	// Wait for goroutine to be ready
	<-ready

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

func TestHandler_DeleteTag_ValidationError(t *testing.T) {
	tagManagerChan := make(chan tag.Manager, 1)
	handler := NewHandler(tagManagerChan)

	router := setupRouter()
	router.DELETE("/tags", handler.DeleteTag)

	// Invalid tag data (invalid EPC)
	tagData := []llrp.TagRecord{
		{PCBits: "3000", EPC: "invalid"},
	}
	jsonData, _ := json.Marshal(tagData)

	req, _ := http.NewRequest("DELETE", "/tags", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response["error"] == nil {
		t.Error("expected error field in response")
	}
}

func TestHandler_DeleteTag_NotFoundError(t *testing.T) {
	tagManagerChan := make(chan tag.Manager, 10)
	handler := NewHandler(tagManagerChan)

	// Set up a goroutine to handle the delete request and respond with empty result (not found)
	ready := make(chan bool)
	go func() {
		close(ready)
		cmd := <-tagManagerChan
		if cmd.Action == tag.DeleteTags {
			// Simulate tag not found by returning empty tags
			cmd.Tags = []*llrp.Tag{}
			tagManagerChan <- cmd
		}
	}()

	// Wait for goroutine to be ready
	<-ready

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

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response["error"] == nil {
		t.Error("expected error field in response")
	}
}

func TestHandler_GetTags_Success(t *testing.T) {
	tagManagerChan := make(chan tag.Manager, 10)
	handler := NewHandler(tagManagerChan)

	// Create tags before starting goroutine so we can handle errors properly
	tag1, err := llrp.NewTag(&llrp.TagRecord{PCBits: "3000", EPC: "001100000111001000100111011000100111111100101110101001001000000000000000000000000001110001101010"})
	if err != nil {
		t.Fatalf("failed to create tag1: %v", err)
	}

	// Set up a goroutine to handle the retrieve request
	// This simulates the tag manager service responding
	ready := make(chan bool)
	go func() {
		close(ready) // Signal that goroutine is ready
		cmd := <-tagManagerChan
		if cmd.Action == tag.RetrieveTags {
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

	// Create a tag to simulate successful addition
	tag1, err := llrp.NewTag(&llrp.TagRecord{PCBits: "3000", EPC: "001100000111001000100111011000100111111100101110101001001000000000000000000000000001110001101010"})
	if err != nil {
		t.Fatalf("failed to create tag1: %v", err)
	}

	// Set up a goroutine to handle the add request and respond
	ready := make(chan bool)
	go func() {
		close(ready)
		cmd := <-tagManagerChan
		if cmd.Action == tag.AddTags {
			// Simulate successful addition by returning the tag
			cmd.Tags = []*llrp.Tag{tag1}
			tagManagerChan <- cmd
		}
	}()

	// Wait for goroutine to be ready
	<-ready

	tagData := []llrp.TagRecord{
		{PCBits: "3000", EPC: "001100000111001000100111011000100111111100101110101001001000000000000000000000000001110001101010"},
	}

	count, err := handler.reqAddTag(tagData)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}
}

func TestHandler_reqAddTag_ValidationError(t *testing.T) {
	tagManagerChan := make(chan tag.Manager, 10)
	handler := NewHandler(tagManagerChan)

	// Invalid tag data (invalid EPC)
	tagData := []llrp.TagRecord{
		{PCBits: "3000", EPC: "invalid"},
	}

	count, err := handler.reqAddTag(tagData)
	if err == nil {
		t.Error("expected validation error, got nil")
	}
	if count != 0 {
		t.Errorf("expected count 0, got %d", count)
	}

	var validationErr *validationError
	if !errors.As(err, &validationErr) {
		t.Errorf("expected validationError, got %T", err)
	}
}

func TestHandler_reqAddTag_DuplicateError(t *testing.T) {
	tagManagerChan := make(chan tag.Manager, 10)
	handler := NewHandler(tagManagerChan)

	// Set up a goroutine to handle the add request and respond with empty result (duplicate)
	ready := make(chan bool)
	go func() {
		close(ready)
		cmd := <-tagManagerChan
		if cmd.Action == tag.AddTags {
			// Simulate duplicate by returning empty tags (tag already exists)
			cmd.Tags = []*llrp.Tag{}
			tagManagerChan <- cmd
		}
	}()

	// Wait for goroutine to be ready
	<-ready

	tagData := []llrp.TagRecord{
		{PCBits: "3000", EPC: "001100000111001000100111011000100111111100101110101001001000000000000000000000000001110001101010"},
	}

	count, err := handler.reqAddTag(tagData)
	if err == nil {
		t.Error("expected duplicateTagError, got nil")
	}
	if count != 0 {
		t.Errorf("expected count 0, got %d", count)
	}

	var duplicateErr *duplicateTagError
	if !errors.As(err, &duplicateErr) {
		t.Errorf("expected duplicateTagError, got %T", err)
	}
}

func TestHandler_PostTag_ValidationError(t *testing.T) {
	tagManagerChan := make(chan tag.Manager, 1)
	handler := NewHandler(tagManagerChan)

	router := setupRouter()
	router.POST("/tags", handler.PostTag)

	// Invalid tag data (invalid EPC)
	tagData := []llrp.TagRecord{
		{PCBits: "3000", EPC: "invalid"},
	}
	jsonData, _ := json.Marshal(tagData)

	req, _ := http.NewRequest("POST", "/tags", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response["error"] == nil {
		t.Error("expected error field in response")
	}
}

func TestHandler_PostTag_DuplicateError(t *testing.T) {
	tagManagerChan := make(chan tag.Manager, 10)
	handler := NewHandler(tagManagerChan)

	// Set up a goroutine to handle the add request and respond with empty result (duplicate)
	ready := make(chan bool)
	go func() {
		close(ready)
		cmd := <-tagManagerChan
		if cmd.Action == tag.AddTags {
			// Simulate duplicate by returning empty tags (tag already exists)
			cmd.Tags = []*llrp.Tag{}
			tagManagerChan <- cmd
		}
	}()

	// Wait for goroutine to be ready
	<-ready

	router := setupRouter()
	router.POST("/tags", handler.PostTag)

	tagData := []llrp.TagRecord{
		{PCBits: "3000", EPC: "001100000111001000100111011000100111111100101110101001001000000000000000000000000001110001101010"},
	}
	jsonData, _ := json.Marshal(tagData)

	req, _ := http.NewRequest("POST", "/tags", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected status %d, got %d", http.StatusConflict, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response["error"] == nil {
		t.Error("expected error field in response")
	}
}

func TestHandler_reqDeleteTag(t *testing.T) {
	tagManagerChan := make(chan tag.Manager, 10)
	handler := NewHandler(tagManagerChan)

	// Create a tag to simulate it being in storage
	tag1, err := llrp.NewTag(&llrp.TagRecord{PCBits: "3000", EPC: "001100000111001000100111011000100111111100101110101001001000000000000000000000000001110001101010"})
	if err != nil {
		t.Fatalf("failed to create tag1: %v", err)
	}

	// Set up a goroutine to handle the delete request and respond
	ready := make(chan bool)
	go func() {
		close(ready)
		cmd := <-tagManagerChan
		if cmd.Action == tag.DeleteTags {
			// Simulate successful deletion by returning the tag
			cmd.Tags = []*llrp.Tag{tag1}
			tagManagerChan <- cmd
		}
	}()

	// Wait for goroutine to be ready
	<-ready

	tagData := []llrp.TagRecord{
		{PCBits: "3000", EPC: "001100000111001000100111011000100111111100101110101001001000000000000000000000000001110001101010"},
	}

	err = handler.reqDeleteTag(tagData)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestHandler_reqDeleteTag_ValidationError(t *testing.T) {
	tagManagerChan := make(chan tag.Manager, 10)
	handler := NewHandler(tagManagerChan)

	// Invalid tag data (invalid EPC)
	tagData := []llrp.TagRecord{
		{PCBits: "3000", EPC: "invalid"},
	}

	err := handler.reqDeleteTag(tagData)
	if err == nil {
		t.Error("expected validation error, got nil")
	}

	var validationErr *validationError
	if !errors.As(err, &validationErr) {
		t.Errorf("expected validationError, got %T", err)
	}
}

func TestHandler_reqDeleteTag_NotFoundError(t *testing.T) {
	tagManagerChan := make(chan tag.Manager, 10)
	handler := NewHandler(tagManagerChan)

	// Set up a goroutine to handle the delete request and respond with empty result (not found)
	ready := make(chan bool)
	go func() {
		close(ready)
		cmd := <-tagManagerChan
		if cmd.Action == tag.DeleteTags {
			// Simulate tag not found by returning empty tags
			cmd.Tags = []*llrp.Tag{}
			tagManagerChan <- cmd
		}
	}()

	// Wait for goroutine to be ready
	<-ready

	tagData := []llrp.TagRecord{
		{PCBits: "3000", EPC: "001100000111001000100111011000100111111100101110101001001000000000000000000000000001110001101010"},
	}

	err := handler.reqDeleteTag(tagData)
	if err == nil {
		t.Error("expected notFoundError, got nil")
	}

	var notFoundErr *notFoundError
	if !errors.As(err, &notFoundErr) {
		t.Errorf("expected notFoundError, got %T", err)
	}
}

func TestHandler_reqRetrieveTag(t *testing.T) {
	tagManagerChan := make(chan tag.Manager, 10)
	handler := NewHandler(tagManagerChan)

	// Create tags before starting goroutine so we can handle errors properly
	tag1, err := llrp.NewTag(&llrp.TagRecord{PCBits: "3000", EPC: "001100000111001000100111011000100111111100101110101001001000000000000000000000000001110001101010"})
	if err != nil {
		t.Fatalf("failed to create tag1: %v", err)
	}
	tag2, err := llrp.NewTag(&llrp.TagRecord{PCBits: "3000", EPC: "001100000111001000100111011000100111111100101110101001001000000000000000000000000001110001101011"})
	if err != nil {
		t.Fatalf("failed to create tag2: %v", err)
	}

	// Set up a goroutine to handle the retrieve request
	// This simulates the tag manager service responding
	ready := make(chan bool)
	go func() {
		close(ready) // Signal that goroutine is ready
		cmd := <-tagManagerChan
		if cmd.Action == tag.RetrieveTags {
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
