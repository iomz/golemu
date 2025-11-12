//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/iomz/golemu/tag"
)

// Server represents the API server
type Server struct {
	handler *Handler
	port    int
}

// NewServer creates a new API server
func NewServer(port int, tagManagerChan chan tag.Manager) *Server {
	return &Server{
		handler: NewHandler(tagManagerChan),
		port:    port,
	}
}

// Start starts the API server
func (s *Server) Start() error {
	r := gin.Default()
	v1 := r.Group("api/v1")
	{
		v1.POST("/tags", s.handler.PostTag)
		v1.DELETE("/tags", s.handler.DeleteTag)
		v1.GET("/tags", s.handler.GetTags)
	}
	return r.Run(":" + strconv.Itoa(s.port))
}
