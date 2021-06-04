package cert

import (
	"github.com/go-acme/lego/v4/lego"
	"testing"
)

const EMAIL = "duyleekun@gmail.com"
const USER_KEY_PATH = "userkey.pem"

func TestRegister(t *testing.T) {
	user, _ := NewMyUserFromEmail(EMAIL)
	err := user.SavePrivateKey(USER_KEY_PATH)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoadKey(t *testing.T) {
	var err error
	user, err := NewMyUserFromFile(EMAIL, USER_KEY_PATH)
	if err != nil {
		t.Fatal(err)
	}

	client, err := NewMyClient(user, lego.LEDirectoryStaging)
	if err != nil {
		t.Fatal(err)
	}

	err = client.ResolveRegistration()
	if err != nil {
		t.Fatal(err)
	}
}

func TestAll(t *testing.T) {

}
