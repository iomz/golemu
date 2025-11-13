//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/fatih/structs"
	"github.com/gin-gonic/gin"
	"github.com/iomz/go-llrp"
	"github.com/iomz/golemu/tag"
	log "github.com/sirupsen/logrus"
)

// validationError represents an error that occurred during tag validation.
type validationError struct {
	message string
	details []string
}

func (e *validationError) Error() string {
	return e.message
}

// notFoundError represents an error when one or more tags are not found in storage.
type notFoundError struct {
	message string
}

func (e *notFoundError) Error() string {
	return e.message
}

// duplicateTagError represents an error when one or more tags already exist.
type duplicateTagError struct {
	message string
}

func (e *duplicateTagError) Error() string {
	return e.message
}

// Handler processes HTTP API requests for tag management operations.
// It provides REST endpoints for adding, deleting, and retrieving tags.
type Handler struct {
	tagManagerChan chan tag.Manager
}

// NewHandler creates a new API handler with the specified tag management channel.
//
// Parameters:
//   - tagManagerChan: Channel for sending tag management commands
func NewHandler(tagManagerChan chan tag.Manager) *Handler {
	return &Handler{
		tagManagerChan: tagManagerChan,
	}
}

// PostTag handles HTTP POST requests to add new tags.
// It expects a JSON array of TagRecord objects in the request body.
// Returns 201 Created on success, 400 Bad Request for invalid JSON or validation errors,
// or 409 Conflict if one or more tags already exist.
func (h *Handler) PostTag(c *gin.Context) {
	var json []llrp.TagRecord
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	count, err := h.reqAddTag(json)
	if err != nil {
		var validationErr *validationError
		var duplicateErr *duplicateTagError
		if errors.As(err, &validationErr) {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.message, "details": validationErr.details})
		} else if errors.As(err, &duplicateErr) {
			c.JSON(http.StatusConflict, gin.H{"error": duplicateErr.message})
		} else {
			log.Errorf("unexpected error in PostTag: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Tags added successfully", "count": count})
}

// DeleteTag handles HTTP DELETE requests to remove tags.
// It expects a JSON array of TagRecord objects in the request body.
// Returns 200 OK on success, 400 Bad Request for invalid JSON or validation errors,
// 404 Not Found if one or more tags do not exist, or 500 Internal Server Error
// for unexpected errors.
func (h *Handler) DeleteTag(c *gin.Context) {
	var json []llrp.TagRecord
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	err := h.reqDeleteTag(json)
	if err != nil {
		var validationErr *validationError
		var notFoundErr *notFoundError
		if errors.As(err, &validationErr) {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.message, "details": validationErr.details})
		} else if errors.As(err, &notFoundErr) {
			c.JSON(http.StatusNotFound, gin.H{"error": notFoundErr.message})
		} else {
			log.Errorf("unexpected error in DeleteTag: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tags deleted successfully"})
}

// GetTags handles HTTP GET requests to retrieve all tags.
// Returns 200 OK with a JSON array of all currently stored tags.
func (h *Handler) GetTags(c *gin.Context) {
	tagList := h.reqRetrieveTag()
	c.JSON(http.StatusOK, tagList)
}

func (h *Handler) reqAddTag(req []llrp.TagRecord) (int, error) {
	validTags := []*llrp.Tag{}
	validationErrors := []string{}
	for i, t := range req {
		tagObj, err := llrp.NewTag(&llrp.TagRecord{
			PCBits: t.PCBits,
			EPC:    t.EPC,
		})
		if err != nil {
			log.Errorf("error creating tag: %v", err)
			validationErrors = append(validationErrors, fmt.Sprintf("tag[%d]: %v", i, err))
			continue
		}

		validTags = append(validTags, tagObj)
	}

	// If there were validation errors, return validationError
	if len(validationErrors) > 0 {
		return 0, &validationError{
			message: "One or more tags failed validation",
			details: validationErrors,
		}
	}

	// Send add commands and wait for responses to check for duplicates
	totalRequested := len(validTags)
	totalAdded := 0

	for _, tagObj := range validTags {
		add := tag.Manager{
			Action: tag.AddTags,
			Tags:   []*llrp.Tag{tagObj},
		}
		h.tagManagerChan <- add
		// Wait for response to check if tag was actually added
		response := <-h.tagManagerChan
		if len(response.Tags) > 0 {
			totalAdded += len(response.Tags)
		}
	}

	// Check if all requested tags were actually added (detect duplicates)
	if totalAdded < totalRequested {
		return totalAdded, &duplicateTagError{
			message: fmt.Sprintf("One or more tags already exist (%d requested, %d added)", totalRequested, totalAdded),
		}
	}

	log.Debugf("add %v", req)
	return totalAdded, nil
}

func (h *Handler) reqDeleteTag(req []llrp.TagRecord) error {
	// First, validate all tags and collect validation errors
	validTags := []*llrp.Tag{}
	validationErrors := []string{}
	for i, t := range req {
		tagObj, err := llrp.NewTag(&llrp.TagRecord{
			PCBits: t.PCBits,
			EPC:    t.EPC,
		})
		if err != nil {
			log.Errorf("error creating tag: %v", err)
			validationErrors = append(validationErrors, fmt.Sprintf("tag[%d]: %v", i, err))
			continue
		}
		validTags = append(validTags, tagObj)
	}

	// If there were validation errors, return validationError
	if len(validationErrors) > 0 {
		return &validationError{
			message: "One or more tags failed validation",
			details: validationErrors,
		}
	}

	// Send delete commands and wait for responses
	totalRequested := len(validTags)
	totalDeleted := 0

	for _, tagObj := range validTags {
		deleteCmd := tag.Manager{
			Action: tag.DeleteTags,
			Tags:   []*llrp.Tag{tagObj},
		}
		h.tagManagerChan <- deleteCmd
		// Wait for response to check if tag was actually deleted
		response := <-h.tagManagerChan
		if len(response.Tags) > 0 {
			totalDeleted += len(response.Tags)
		}
	}

	// Check if all requested tags were actually deleted
	if totalDeleted < totalRequested {
		return &notFoundError{
			message: fmt.Sprintf("One or more tags not found (%d requested, %d deleted)", totalRequested, totalDeleted),
		}
	}

	log.Debugf("delete %v", req)
	return nil
}

func (h *Handler) reqRetrieveTag() []map[string]interface{} {
	retrieve := tag.Manager{
		Action: tag.RetrieveTags,
		Tags:   []*llrp.Tag{},
	}
	h.tagManagerChan <- retrieve
	retrieve = <-h.tagManagerChan
	var tagList []map[string]interface{}
	for _, tagObj := range retrieve.Tags {
		t := structs.Map(llrp.NewTagRecord(*tagObj))
		tagList = append(tagList, t)
	}
	log.Debugf("retrieve: %v", tagList)
	return tagList
}
