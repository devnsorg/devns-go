package main

import (
	"flag"
	"fmt"
	"github.com/ipTLS/ipTLS/wgclient/mywgclient"
	"github.com/mdp/qrterminal/v3"
	"golang.zx2c4.com/wireguard/device"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var portF = flag.Int("port", 8888, "Port for DNS server to listen to")
var rootDomainF = flag.String("domain", "example.com", "[MUST CHANGE] Base domain for DNS resolution")
var apiEndpoint = flag.String("api-endpoint", "http://localhost:9999/wg", "wg api endpoint")
var qrOnly = flag.Bool("qr", false, "Set true if print QR only. Default connect to WG")

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
	var wgClient *mywgclient.WGClient
	go func() {
		select {
		case sig := <-signalChan:
			logger.Errorf("signalChan %#v", sig)
			wgClient.StopClient()
			os.Exit(0)
		case err := <-errsChan:
			logger.Errorf("ERRSCHAN %#v", err)
			wgClient.StopClient()
			os.Exit(2)
		}
	}()

	resp, err := http.Get(*apiEndpoint)
	if err != nil {
		log.Printf("ERROR %e\n", err)
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("ERROR %e\n", err)
	}
	if *qrOnly {
		qrterminal.Generate(string(respBytes), qrterminal.L, os.Stdout)
	} else {
		wgClient = mywgclient.NewWGClient(string(respBytes), logger, errsChan)
		wgChan := wgClient.StartServer()
		logger.Verbosef("WG STARTED")

		select {
		case <-wgChan:
		}
	}
}
