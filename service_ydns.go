package main

import (
	"errors"
	"strings"
)

type yDNSService Ddns

func (s *yDNSService) getDomain() string {
	return s.Domain
}

func (s *yDNSService) updateIP() error {
	content, err := GetResponse("https://ydns.eu/api/v1/update/?host=" + s.Domain, s.UserName, s.Password)
	if err != nil {
		return err
	}

	if !strings.HasPrefix(content, "ok") {
		return errors.New(content)
	}

	return nil
}
