package main

import (
	"errors"
	"strings"
)

type nsupdateService Ddns

func (s *nsupdateService) getDomain() string {
	return s.Domain
}

func (s *nsupdateService) updateIP() error {
	content, err := GetResponse("https://ipv4.nsupdate.info/nic/update", s.Domain, s.Password)
	if err != nil {
		return err
	}

	if !strings.HasPrefix(content, "nochg") && !strings.HasPrefix(content, "good") {
		return errors.New(content)
	}

	return nil
}
