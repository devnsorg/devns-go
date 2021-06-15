package main

import (
	"flag"
	"fmt"
	"github.com/ipTLS/ipTLS/wgapi/myhttpserver"
	"github.com/ipTLS/ipTLS/wgapi/mywgserver"
	"golang.zx2c4.com/wireguard/device"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var portF = flag.Int("port", 8888, "Port for DNS server to listen to")
var rootDomainF = flag.String("domain", "example.com", "[MUST CHANGE] Base domain for DNS resolution")
var wgEndpointF = flag.String("wg-endpoint", "192.168.0.11:51820", "[MUST CHANGE] Base domain for DNS resolution")
var helpF = flag.Bool("h", false, "Print this help")

func main() {

	log.SetFlags(log.Lshortfile)
	flag.Parse()

	logger := &device.Logger{
		Verbosef: func(format string, args ...interface{}) {
			log.SetPrefix("[VERBOSE] ")
			log.Output(2, fmt.Sprintf(format, args...))
		},
		Errorf: func(format string, args ...interface{}) {
			log.SetPrefix("[ERROR] ")
			log.Output(2, fmt.Sprintf(format, args...))
		},
	}

	errsChan := make(chan error)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM)
	signal.Notify(signalChan, os.Interrupt)
	var wgServer *mywgserver.WGServer
	go func() {
		select {
		case sig := <-signalChan:
			logger.Errorf("signalChan %#v", sig)
			wgServer.StopServer()
			os.Exit(0)
		case err := <-errsChan:
			logger.Errorf("ERRSCHAN %#v", err)
			wgServer.StopServer()
			os.Exit(2)
		}
	}()
	wgServer = mywgserver.NewWGServer(*wgEndpointF, "10.44.0.1/23", logger, errsChan)
	wgChan := wgServer.StartServer()
	logger.Verbosef("WG STARTED")
	httpServer := myhttpserver.NewHTTPServer(9999, wgServer, logger, errsChan)
	httpChan := httpServer.StartServer()
	logger.Verbosef("HTTP STARTED")

	select {
	case <-wgChan:
	case <-httpChan:
	}
}
