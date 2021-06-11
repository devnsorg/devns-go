package myhttpserver

import (
	"errors"
	"fmt"
	"github.com/go-acme/lego/v4/log"
	"github.com/ipTLS/ipTLS/tlsapi/cert"
	"github.com/miekg/dns"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
}

func (h Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	var err error
	questionBytes, _ := ioutil.ReadAll(request.Body)
	questionString := string(questionBytes)
	log.Infof("Question: '%s'", questionString)

	matches := strings.Split(questionString, " ")
	typeInt, _ := strconv.Atoi(matches[1])
	classInt, _ := strconv.Atoi(matches[2])
	question := dns.Question{
		Name:   matches[0],
		Qtype:  uint16(typeInt),
		Qclass: uint16(classInt),
	}
	var rr dns.RR
	switch question.Qtype {
	case dns.TypeTXT:
		if len(cert.GetChallenge(strings.ToLower(question.Name))) == 0 {
			err = errors.New("TXT challenge response not found")
			rr = &dns.TXT{
				Hdr: dns.RR_Header{
					Name:     question.Name,
					Rrtype:   question.Qtype,
					Class:    question.Qclass,
					Ttl:      60,
					Rdlength: 0,
				},
				Txt: []string{"NOT FOUND"},
			}
		} else {
			rr = &dns.TXT{
				Hdr: dns.RR_Header{
					Name:     question.Name,
					Rrtype:   question.Qtype,
					Class:    question.Qclass,
					Ttl:      60,
					Rdlength: 0,
				},
				Txt: []string{cert.GetChallenge(strings.ToLower(question.Name))},
			}
		}
	}
	if err != nil {
		log.Infof("ERROR : %s", err)
	}
	if rr != nil {
		_, err := writer.Write([]byte(rr.String()))

		if err != nil {
			log.Infof("ERROR : %s", err)
		}
	} else {
		writer.WriteHeader(404)
	}

}

func StartServer(port int) {
	http.ListenAndServe(fmt.Sprintf(":%d", port), Handler{})
}
