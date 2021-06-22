package main

import (
	"flag"
	"fmt"
	"github.com/devnsorg/devns-go/wgclient/mywgclient"
	"github.com/mdp/qrterminal/v3"
	"golang.zx2c4.com/wireguard/device"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
)

var apiEndpoint = flag.String("api-endpoint", "http://localhost:9999/wg", "wg api endpoint")
var subdomain = flag.String("subdomain", "", "subdomain for NS")
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
	endpointURL, err := url.Parse(*apiEndpoint)
	if err != nil {
		log.Fatalf("ERROR %e\n", err)
	}
	q := endpointURL.Query()
	q.Add("subdomain", *subdomain)
	endpointURL.RawQuery = q.Encode()

	log.Printf("REQUEST %s", endpointURL.String())
	resp, err := http.Get(endpointURL.String())
	if err != nil {
		log.Fatalf("http.Get %e\n", err)
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("ioutil.ReadAll %e\n", err)
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
