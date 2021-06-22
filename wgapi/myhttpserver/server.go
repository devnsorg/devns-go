package myhttpserver

import (
	"errors"
	"fmt"
	"github.com/devnsorg/devns-go/wgapi/mywgserver"
	"github.com/gorilla/mux"
	"github.com/miekg/dns"
	"golang.zx2c4.com/wireguard/device"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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
		var err error
		questionBytes, _ := ioutil.ReadAll(request.Body)
		questionString := string(questionBytes)

		matches := strings.Split(questionString, " ")
		typeInt, _ := strconv.Atoi(matches[1])
		classInt, _ := strconv.Atoi(matches[2])
		question := dns.Question{
			Name:   matches[0],
			Qtype:  uint16(typeInt),
			Qclass: uint16(classInt),
		}

		var rr dns.RR
		switch question.Qtype {
		case dns.TypeA:
			if val := h.server.GetPeerAddressFor(strings.ToLower(question.Name)); val != nil {
				rr = &dns.A{
					Hdr: dns.RR_Header{
						Name:     question.Name,
						Rrtype:   question.Qtype,
						Class:    question.Qclass,
						Ttl:      60,
						Rdlength: 0,
					},
					A: val,
				}
			} else {
				err = errors.New("A not found")

			}
		}
		if err != nil {
			h.logger.Errorf("ERROR : %s", err)
			writer.WriteHeader(404)
		} else {
			if rr != nil {
				_, err := writer.Write([]byte(rr.String()))

				if err != nil {
					h.logger.Errorf("ERROR : %s", err)
				}
			}

		}
	})
	r.HandleFunc("/wg", func(writer http.ResponseWriter, request *http.Request) {
		requestURL, _ := url.Parse(request.RequestURI)
		subdomain := requestURL.Query().Get("subdomain")
		writer.Write(h.server.AddClientPeer(subdomain))
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
