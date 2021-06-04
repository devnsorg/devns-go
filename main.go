package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/miekg/dns"
	"log"
	"strconv"
	"strings"
)

var portF = flag.Int("port", 53, "Port for DNS server to listen to")
var v4RootDomainF = flag.String("v4Domain", "", "Base domain for DNS resolution")
var v6RootDomainF = flag.String("v6Domain", "", "Base domain for DNS resolution")

func parseQuery(rootDomain string, dnsMessage *dns.Msg) {
	for _, question := range dnsMessage.Question {
		log.Printf("Query for %s\n", question.String())
		ip := strings.ReplaceAll(question.Name, "."+rootDomain+".", "")
		var err error
		var rr dns.RR
		var recordType string
		var localIP string
		switch question.Qtype {
		case dns.TypeAAAA:
			recordType = "AAAA"
			localIP = "::1"
			ip = strings.ReplaceAll(ip, "-", ":")
		case dns.TypeA:
			recordType = "A"
			localIP = "127.0.0.1"
			ip = strings.ReplaceAll(ip, "-", ".")
		default:
			err = errors.New(question.String())
		}
		if err == nil {
			if ip == "localhost" {
				ip = localIP
			}
			rr, err = dns.NewRR(fmt.Sprintf("%s %s %s", question.Name, recordType, ip))
		}
		if err == nil {
			log.Printf("Answer for %s\n", rr.String())
			dnsMessage.Answer = append(dnsMessage.Answer, rr)
		}
		if err != nil {
			log.Println("%@", err)
		}
	}
}

func handleDnsRequest(rootDomain string) func(dns.ResponseWriter, *dns.Msg) {
	return func(responseWriter dns.ResponseWriter, dnsMessage *dns.Msg) {
		newDnsMessage := new(dns.Msg)
		newDnsMessage.SetReply(dnsMessage)
		newDnsMessage.Compress = false

		switch dnsMessage.Opcode {
		case dns.OpcodeQuery:
			parseQuery(rootDomain, newDnsMessage)

		}
		responseWriter.WriteMsg(newDnsMessage)
	}
}

func main() {
	flag.Parse()
	// attach request handler func
	v4RootDomain := *v4RootDomainF + "."
	v6RootDomain := *v6RootDomainF + "."
	shouldPrintUsage := true

	if len(v4RootDomain) > 1 {
		dns.HandleFunc(v4RootDomain, handleDnsRequest(v4RootDomain))
		shouldPrintUsage = false
	}

	if len(v6RootDomain) > 1 {
		dns.HandleFunc(v6RootDomain, handleDnsRequest(v6RootDomain))
		shouldPrintUsage = false
	}

	if shouldPrintUsage {
		flag.Usage()
		return
	}

	// start server
	port := *portF
	server := &dns.Server{Addr: ":" + strconv.Itoa(port), Net: "udp"}
	log.Printf("Starting at %d for v4 %s and v6 %s\n", port, v4RootDomain, v6RootDomain)
	err := server.ListenAndServe()
	defer server.Shutdown()
	if err != nil {
		log.Fatalf("Failed to start server: %s\n ", err.Error())
	}
}
