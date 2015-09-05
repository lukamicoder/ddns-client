package main

import (
	"errors"
	"strings"
)

type duckDNSService Ddns

func (s *duckDNSService) getDomain() string {
	return s.Domain
}

func (s *duckDNSService) updateIP() error {
	pos := strings.Index(s.Domain, ".")
	if pos < 1 {
		return errors.New("Incorrect domain.")
	}

	host := s.Domain[0:pos]
	url := "https://www.duckdns.org/update?domains=" + host + "&token=" + s.Token

	content, err := GetResponse(url, "", "")
	if err != nil {
		return err
	}

	if content == "KO" {
		return errors.New("Update failed.")
	}

	return nil
}
