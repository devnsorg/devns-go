package cert

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"github.com/devnsorg/devns-go/tlsapi/util"
	"github.com/go-acme/lego/v4/registration"
	"log"
)

type MyUser struct {
	email        string
	registration *registration.Resource
	key          *ecdsa.PrivateKey
}

func (u *MyUser) GetEmail() string {
	return u.email
}
func (u MyUser) GetRegistration() *registration.Resource {
	return u.registration
}
func (u *MyUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

func (u *MyUser) SavePrivateKey(filename string) error {
	return util.KeyToFile(u.key, filename)
}

func NewMyUserFromEmail(email string) (*MyUser, error) {
	// Create a user. New accounts need an email and private key to start.
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	if err != nil {
		log.Fatalln(err)
	}

	myUser := MyUser{
		email: email,
		key:   privateKey,
	}

	return &myUser, err
}

func NewMyUserFromFile(email string, filename string) (*MyUser, error) {
	privateKey, err := util.FileToKey(filename)

	if err != nil {
		log.Fatalln(err)
	}

	myUser := MyUser{
		email: email,
		key:   privateKey,
	}

	return &myUser, err
}
