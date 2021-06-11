package mydns

import (
	"bytes"
	"fmt"
	"github.com/miekg/dns"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
)

func parseQuery(rootDomain string, nameserver string, soaEmail string, apiEndpoints []string, dnsMessage *dns.Msg) {
	for _, question := range dnsMessage.Question {
		log.Printf("Query for %s\n", question.String())
		subdomain := strings.ReplaceAll(strings.ToLower(question.Name), "."+rootDomain, "")
		var rr dns.RR = nil
		hdr := dns.RR_Header{
			Name:     question.Name,
			Rrtype:   question.Qtype,
			Class:    question.Qclass,
			Ttl:      60,
			Rdlength: 0,
		}
		switch question.Qtype {
		case dns.TypeAAAA:
			var ip net.IP
			if subdomain == "localhost" {
				ip = net.ParseIP(strings.ReplaceAll("::1", "-", "."))
			} else {
				ip = net.ParseIP(strings.ReplaceAll(strings.ReplaceAll(subdomain, "-", ":"), "-", "."))
			}
			if ip != nil {
				rr = &dns.AAAA{
					Hdr:  hdr,
					AAAA: ip,
				}
			}
		case dns.TypeA:
			var ip net.IP
			if subdomain == "localhost" {
				ip = net.ParseIP(strings.ReplaceAll("127.0.0.1", "-", "."))

			} else {
				ip = net.ParseIP(strings.ReplaceAll(subdomain, "-", "."))
			}
			if ip != nil {
				rr = &dns.A{
					Hdr: hdr,
					A:   ip,
				}
			}
		case dns.TypeSOA:
			rr = &dns.SOA{
				Hdr:     hdr,
				Ns:      nameserver,
				Mbox:    soaEmail,
				Serial:  1,
				Refresh: 10000,
				Retry:   2400,
				Expire:  604800,
				Minttl:  3600,
			}
		case dns.TypeNS:
			rr = &dns.NS{
				Hdr: hdr,
				Ns:  nameserver,
			}
		}
		if rr == nil {
			for _, apiEndpoint := range apiEndpoints {
				resp, err := http.Post(apiEndpoint, "application/json", bytes.NewBuffer([]byte(fmt.Sprintf("%s %d %d", question.Name, question.Qtype, question.Qclass))))
				if err != nil {
					log.Printf("ERROR %e\n", err)
				}

				respBytes, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Printf("ERROR %e\n", err)
				}
				log.Printf("ANSWERS FROM API %s", string(respBytes))
				rr, err = dns.NewRR(string(respBytes))
				if err != nil {
					log.Printf("ERROR %e\n", err)
				}
				if rr != nil {
					break
				}
			}

		}

		if rr != nil {
			log.Printf("Answer for \"%s\" is \"%s\"\n", question.String(), rr.String())
			dnsMessage.Answer = append(dnsMessage.Answer, rr)
		}
	}
}

func HandleDnsRequest(rootDomain string, nameserver string, soaEmail string, apiEndpoints []string) func(dns.ResponseWriter, *dns.Msg) {
	return func(responseWriter dns.ResponseWriter, dnsMessage *dns.Msg) {
		newDnsMessage := new(dns.Msg)
		newDnsMessage.SetReply(dnsMessage)
		newDnsMessage.Compress = false

		switch dnsMessage.Opcode {
		case dns.OpcodeQuery:
			parseQuery(rootDomain, nameserver, soaEmail, apiEndpoints, newDnsMessage)

		}
		responseWriter.WriteMsg(newDnsMessage)
	}
}
