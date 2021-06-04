package cert

import (
	"github.com/go-acme/lego/v4/lego"
	"io/ioutil"
	"log"
)

func StartCertFlow(domain string, email string, dryrun bool) {
	var err error
	var client *MyClient
	user, err := NewMyUserFromEmail(email)
	if err != nil {
		log.Fatalln(err)
	}
	if dryrun {
		client, err = NewMyClient(user, lego.LEDirectoryStaging)
	} else {
		client, err = NewMyClient(user, lego.LEDirectoryProduction)
	}

	if err != nil {
		log.Fatalln(err)
	}

	err = client.Register()
	if err != nil {
		log.Fatalln(err)
	}

	cert, err := client.GetCert(domain)
	if err != nil {
		log.Fatalln(err)
	}
	err = ioutil.WriteFile("cert.crt", cert.Certificate, 0644)
	if err != nil {
		log.Fatalln(err)
	}
	err = ioutil.WriteFile("cert.key", cert.PrivateKey, 0644)
	if err != nil {
		log.Fatalln(err)
	}
	err = ioutil.WriteFile("ca.crt", cert.IssuerCertificate, 0644)
	if err != nil {
		log.Fatalln(err)
	}
}
