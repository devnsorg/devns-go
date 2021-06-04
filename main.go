package main

import (
	"flag"
	"fmt"
	"github.com/miekg/dns"
	"log"
	"strconv"
	"strings"
)

var portF = *flag.Int("port", 53, "Port for DNS server to listen to")
var v4RootDomainF = *flag.String("v4Domain", "ipv4.iptls.com", "Base domain for DNS resolution") + "."
var v6RootDomainF = *flag.String("v6Domain", "ipv6.iptls.com", "Base domain for DNS resolution") + "."

func parseQuery(m *dns.Msg) {
	for _, q := range m.Question {
		log.Printf("Query for %s\n", q.Name)
		ip := strings.ReplaceAll(q.Name, "."+v6RootDomainF, "")
		ip = strings.ReplaceAll(ip, "."+v4RootDomainF, "")
		var err error
		var rr dns.RR

		switch ip {
		case "localhost":
			switch q.Qtype {
			case dns.TypeAAAA:
				rr, err = dns.NewRR(fmt.Sprintf("%s AAAA ::1", q.Name))
			case dns.TypeA:
				rr, err = dns.NewRR(fmt.Sprintf("%s A 127.0.0.1", q.Name))
			}
		default:
			switch q.Qtype {
			case dns.TypeAAAA:
				ip = strings.ReplaceAll(ip, "-", ":")
				rr, err = dns.NewRR(fmt.Sprintf("%s AAAA %s", q.Name, ip))
			case dns.TypeA:
				ip = strings.ReplaceAll(ip, "-", ".")
				rr, err = dns.NewRR(fmt.Sprintf("%s A %s", q.Name, ip))
			}
		}

		if err == nil {
			log.Printf("Answer for %s\n", rr.String())
			m.Answer = append(m.Answer, rr)
		} else {
			log.Println("%@", err)
		}
	}
}

func handleDnsRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	switch r.Opcode {
	case dns.OpcodeQuery:
		parseQuery(m)
	}

	w.WriteMsg(m)
}

func main() {
	// attach request handler func
	dns.HandleFunc(v4RootDomainF, handleDnsRequest)
	dns.HandleFunc(v6RootDomainF, handleDnsRequest)

	// start server
	port := portF
	server := &dns.Server{Addr: ":" + strconv.Itoa(portF), Net: "udp"}
	log.Printf("Starting at %d\n", port)
	err := server.ListenAndServe()
	defer server.Shutdown()
	if err != nil {
		log.Fatalf("Failed to start server: %s\n ", err.Error())
	}
}
