package cert

import (
	"github.com/go-acme/lego/v4/challenge/dns01"
	"log"
)

var challenges = make(map[string]string)

func GetChallenge(domain string) string {
	return challenges[domain]
}

type DNSProviderBestDNS struct {
}

func NewDNSProviderBestDNS() (*DNSProviderBestDNS, error) {
	return &DNSProviderBestDNS{}, nil
}

func (d *DNSProviderBestDNS) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	log.Println(fqdn, value)
	challenges[fqdn] = value
	return nil
}

func (d *DNSProviderBestDNS) CleanUp(domain, token, keyAuth string) error {
	// clean up any state you created in Present, like removing the TXT record
	delete(challenges, domain)
	return nil
}
