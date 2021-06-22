package main

import (
	"flag"
	"github.com/devnsorg/devns-go/dnsserver/mydns"
	"github.com/miekg/dns"
	"log"
	"os"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return "my string representation"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var portF = flag.Int("port", 53, "Port for DNS server to listen to")
var rootDomainF = flag.String("domain", "example.com", "[MUST CHANGE] Base domain for DNS resolution")
var nameserverF = flag.String("nameserver", "ns.example.com.", "[MUST CHANGE] Primary NS for SOA must end with period(.)")
var soaEmailF = flag.String("soa-email", "john\\n.doe.example.com.", "Email for SOA must end with period(.)")
var helpF = flag.Bool("h", false, "Print this help")

func main() {
	var apiEndpoints arrayFlags
	flag.Var(&apiEndpoints, "api-endpoint", "Specify multiple value for calling multiple API to get result")

	log.SetFlags(log.LstdFlags | log.Llongfile)
	flag.Parse()
	if *helpF || len(os.Args[1:]) == 0 {
		flag.Usage()
		return
	}

	rootDomain := *rootDomainF + "."
	dns.HandleFunc(rootDomain, mydns.HandleDnsRequest(rootDomain, *nameserverF, *soaEmailF, apiEndpoints))

	// start DNS server
	go mydns.StartServer(*portF)

	select {}
}
