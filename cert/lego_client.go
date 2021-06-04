package cert

import (
	"errors"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
	"log"
)

type MyClient struct {
	user     *MyUser
	client   *lego.Client
	cADirURL string
	resolved bool
}

func NewMyClient(myUser *MyUser, cADirURL string) (*MyClient, error) {
	config := lego.NewConfig(myUser)

	// This CA URL is configured for a local dev instance of Boulder running in Docker in a VM.
	config.CADirURL = cADirURL
	config.Certificate.KeyType = certcrypto.RSA2048

	client, err := lego.NewClient(config)
	if err != nil {
		log.Fatalln(err)
	}

	bestDNS, err := NewDNSProviderBestDNS()
	err = client.Challenge.SetDNS01Provider(bestDNS)
	if err != nil {
		log.Fatalln(err)
	}

	return &MyClient{user: myUser, client: client, cADirURL: cADirURL, resolved: false}, err
}

func (c *MyClient) Register() error {
	// New users will need to register
	log.Println("client.Registration.Register")
	reg, err := c.client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		log.Fatalln(err)
	} else {
		c.user.registration = reg
		c.resolved = true
	}

	return err
}

func (c *MyClient) ResolveRegistration() error {
	key, err := c.client.Registration.ResolveAccountByKey()
	if err != nil {
		log.Fatalln(err)
	} else {
		c.user.registration = key
		c.resolved = true
	}
	return err
}

func (c *MyClient) GetCert(domain string) (*certificate.Resource, error) {
	if !c.resolved {
		return nil, errors.New("must Resolve or Register before call API")
	}

	request := certificate.ObtainRequest{
		Domains: []string{domain},
		Bundle:  true,
	}
	certificates, err := c.client.Certificate.Obtain(request)
	if err != nil {
		log.Fatalln(err)
	}

	//fmt.Printf("%#v\n", certificates)

	// Each certificate comes back with the cert bytes, the bytes of the client's
	// private key, and a certificate URL. SAVE THESE TO DISK.
	return certificates, err
}
