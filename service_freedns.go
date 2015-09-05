package main

import (
	"errors"
)

type freeDNSService Ddns

func (s *freeDNSService) getDomain() string {
	return s.Domain
}

func (s *freeDNSService) updateIP() error {
	url := "http://freedns.afraid.org/dynamic/update.php?" + s.Token

	content, err := GetResponse(url, "", "")
	if err != nil {
		return err
	}

	if content[0:5] == "ERROR" {
		return errors.New(content)
	}

	return nil
}
