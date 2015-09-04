package main

import (
	"errors"
	"strings"
)

type IPDNSService Ddns

func (s *IPDNSService) GetDomain() string {
	return s.Domain
}

func (s *IPDNSService) UpdateIP() error {
	content, err := GetResponse("http://update.ipdns.hu/update?hostname=" + s.Domain, s.Account, s.Password)
	if err != nil {
		return err
	}

	if !strings.HasPrefix(content, "good") {
		return errors.New(content)
	}

	return nil
}
