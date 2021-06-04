package cert

import (
	"errors"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"log"
)

type DNSProviderBestDNS struct {
}

func NewDNSProviderBestDNS() (*DNSProviderBestDNS, error) {
	return &DNSProviderBestDNS{}, nil
}

func (d *DNSProviderBestDNS) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	log.Println(fqdn, value)
	// make API request to set a TXT record on fqdn with value and TTL
	//return errors.New("TODO Present")
	return nil
}

func (d *DNSProviderBestDNS) CleanUp(domain, token, keyAuth string) error {
	// clean up any state you created in Present, like removing the TXT record
	return errors.New("TODO CleanUp")
}
