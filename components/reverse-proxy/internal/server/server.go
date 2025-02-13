package server

import (
	"net/http"

	"go.uber.org/zap"
)

// Server represents a general server instance with logs
type Server struct {
	HTTPServer *http.Server
	Log        *zap.SugaredLogger
}

// Serve wraps the default ListenAndServe method and enriches it with error handling and multi-threading support
func (s *Server) Serve(stop chan bool) {
	err := s.HTTPServer.ListenAndServe()
	if err != nil {
		stop <- true
		s.Log.Errorf("error while serving: %v", err)
	}
}
