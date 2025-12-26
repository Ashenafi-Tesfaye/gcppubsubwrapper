package server

import (
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, srv *Server) {
	mux.HandleFunc("/publish", srv.HandlePublish)
}
