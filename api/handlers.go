//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package api

import (
	"net/http"

	"github.com/fatih/structs"
	"github.com/gin-gonic/gin"
	"github.com/iomz/go-llrp"
	"github.com/iomz/golemu/tag"
	log "github.com/sirupsen/logrus"
)

// Handler handles API requests
type Handler struct {
	tagManagerChan chan tag.Manager
}

// NewHandler creates a new API handler
func NewHandler(tagManagerChan chan tag.Manager) *Handler {
	return &Handler{
		tagManagerChan: tagManagerChan,
	}
}

// PostTag handles tag addition requests
func (h *Handler) PostTag(c *gin.Context) {
	var json []llrp.TagRecord
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	if res := h.reqAddTag(json); res == "error" {
		c.JSON(http.StatusConflict, gin.H{"error": "One or more tags already exist"})
	} else {
		c.JSON(http.StatusCreated, gin.H{"message": "Tags added successfully"})
	}
}

// DeleteTag handles tag deletion requests
func (h *Handler) DeleteTag(c *gin.Context) {
	var json []llrp.TagRecord
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	if res := h.reqDeleteTag(json); res == "error" {
		c.JSON(http.StatusNotFound, gin.H{"error": "One or more tags not found"})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Tags deleted successfully"})
	}
}

// GetTags handles tag retrieval requests
func (h *Handler) GetTags(c *gin.Context) {
	tagList := h.reqRetrieveTag()
	c.JSON(http.StatusOK, tagList)
}

func (h *Handler) reqAddTag(req []llrp.TagRecord) string {
	validTags := []*llrp.Tag{}
	for _, t := range req {
		tagObj, err := llrp.NewTag(&llrp.TagRecord{
			PCBits: t.PCBits,
			EPC:    t.EPC,
		})
		if err != nil {
			log.Errorf("error creating tag: %v", err)
			return "error"
		}

		validTags = append(validTags, tagObj)
	}

	for _, tagObj := range validTags {
		add := tag.Manager{
			Action: tag.AddTags,
			Tags:   []*llrp.Tag{tagObj},
		}
		h.tagManagerChan <- add
	}

	log.Debugf("add %v", req)
	return "add"
}

func (h *Handler) reqDeleteTag(req []llrp.TagRecord) string {
	hasError := false
	for _, t := range req {
		tagObj, err := llrp.NewTag(&llrp.TagRecord{
			PCBits: t.PCBits,
			EPC:    t.EPC,
		})
		if err != nil {
			log.Errorf("error creating tag: %v", err)
			hasError = true
			continue
		}

		deleteCmd := tag.Manager{
			Action: tag.DeleteTags,
			Tags:   []*llrp.Tag{tagObj},
		}
		h.tagManagerChan <- deleteCmd
	}

	if hasError {
		return "error"
	}
	log.Debugf("delete %v", req)
	return "delete"
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
