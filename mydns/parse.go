package mydns

import (
	"errors"
	"github.com/ipTLS/dnsserver/cert"
	"github.com/miekg/dns"
	"log"
	"net"
	"strings"
)

func parseQuery(rootDomain string, nameserver string, soaEmail string, dnsMessage *dns.Msg) {
	for _, question := range dnsMessage.Question {
		log.Printf("Query for %s\n", question.String())
		subdomain := strings.ReplaceAll(strings.ToLower(question.Name), "."+rootDomain, "")
		var err error
		var rr dns.RR
		hdr := dns.RR_Header{
			Name:     question.Name,
			Rrtype:   question.Qtype,
			Class:    question.Qclass,
			Ttl:      60,
			Rdlength: 0,
		}
		switch question.Qtype {
		case dns.TypeAAAA:
			if subdomain == "localhost" {
				rr = &dns.AAAA{
					Hdr:  hdr,
					AAAA: net.ParseIP(strings.ReplaceAll("::1", "-", ".")),
				}
			} else {
				rr = &dns.AAAA{
					Hdr:  hdr,
					AAAA: net.ParseIP(strings.ReplaceAll(strings.ReplaceAll(subdomain, "-", ":"), "-", ".")),
				}
			}

		case dns.TypeA:
			if subdomain == "localhost" {
				rr = &dns.A{
					Hdr: hdr,
					A:   net.ParseIP(strings.ReplaceAll("127.0.0.1", "-", ".")),
				}
			} else {
				rr = &dns.A{
					Hdr: hdr,
					A:   net.ParseIP(strings.ReplaceAll(subdomain, "-", ".")),
				}
			}
		case dns.TypeTXT:
			if len(cert.GetChallenge(strings.ToLower(question.Name))) == 0 {
				err = errors.New("TXT challenge response not found")
			} else {
				rr = &dns.TXT{
					Hdr: hdr,
					Txt: []string{cert.GetChallenge(strings.ToLower(question.Name))},
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
		default:
			err = errors.New(question.String())
		}
		if err == nil {
			log.Printf("Answer for \"%s\" is \"%s\"\n", question.String(), rr.String())
			dnsMessage.Answer = append(dnsMessage.Answer, rr)
		}
		if err != nil {
			log.Printf("ERROR %e\n", err)
		}
	}
}

func HandleDnsRequest(rootDomain string, nameserver string, soaEmail string) func(dns.ResponseWriter, *dns.Msg) {
	return func(responseWriter dns.ResponseWriter, dnsMessage *dns.Msg) {
		newDnsMessage := new(dns.Msg)
		newDnsMessage.SetReply(dnsMessage)
		newDnsMessage.Compress = false

		switch dnsMessage.Opcode {
		case dns.OpcodeQuery:
			parseQuery(rootDomain, nameserver, soaEmail, newDnsMessage)

		}
		responseWriter.WriteMsg(newDnsMessage)
	}
}
