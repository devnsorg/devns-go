package main

import (
	"flag"
	"github.com/ipTLS/ipTLS/tlsapi/myhttpserver"
	"log"
	"os"
)

var portF = flag.Int("port", 8888, "Port for DNS server to listen to")
var rootDomainF = flag.String("domain", "example.com", "[MUST CHANGE] Base domain for DNS resolution")
var withTlsF = flag.Bool("tls", false, "Turn on TLS mode")
var tlsEmailF = flag.String("tls-email", "john.doe@example.com", "[MUST CHANGE] Email for letsencrypt registration")
var tlsDryRunF = flag.Bool("tls-dryrun", false, "Set to use STAGING ACME Directory")
var helpF = flag.Bool("h", false, "Print this help")

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	flag.Parse()
	if *helpF || len(os.Args[1:]) == 0 {
		flag.Usage()
		return
	}

	myhttpserver.StartServer(*portF)

	//if *withTlsF {
	//	if len(*tlsEmailF) > 0 {
	//		if len(*rootDomainF) > 0 {
	//			go cert.StartCertFlow("*."+*rootDomainF, *tlsEmailF, *tlsDryRunF)
	//		}
	//	}
	//}

	select {}
}
