//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/iomz/golemu/tag"
)

// Server provides an HTTP API server for tag management operations.
// It exposes REST endpoints for adding, deleting, and retrieving tags.
type Server struct {
	handler *Handler
	port    int
}

// NewServer creates and initializes a new API server.
//
// Parameters:
//   - port: Port number to listen on
//   - tagManagerChan: Channel for tag management operations
func NewServer(port int, tagManagerChan chan tag.Manager) *Server {
	return &Server{
		handler: NewHandler(tagManagerChan),
		port:    port,
	}
}

// Start starts the HTTP API server and begins listening for requests.
// It registers routes for POST /api/v1/tags, DELETE /api/v1/tags, and GET /api/v1/tags.
// The server runs until an error occurs or it is stopped.
//
// Returns an error if the server cannot start or encounters a fatal error.
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
