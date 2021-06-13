package myhttpserver

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/ipTLS/ipTLS/wgapi/mywgserver"
	"golang.zx2c4.com/wireguard/device"
	"net/http"
)

type HTTPServer struct {
	server *mywgserver.WGServer
	errs   chan error
	logger *device.Logger
	port   int
}

func NewHTTPServer(port int, server *mywgserver.WGServer, logger *device.Logger, errs chan error) *HTTPServer {
	return &HTTPServer{port: port, server: server, logger: logger, errs: errs}
}

func (h *HTTPServer) StartServer() chan error {
	closed := make(chan error)
	r := mux.NewRouter()
	r.HandleFunc("/dns", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(204)
	})
	r.HandleFunc("/wg", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write(h.server.AddClientPeer())
	})
	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%d", h.port), r)
		if err != nil {
			h.logger.Errorf("ERROR %#v", err)
			h.errs <- err
		}
		closed <- err
	}()
	return closed
}
