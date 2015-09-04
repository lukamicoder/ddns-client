package main

import (
	"errors"
)

type FreeDNSService Ddns

func (s *FreeDNSService) GetDomain() string {
	return s.Domain
}

func (s *FreeDNSService) UpdateIP() error {
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
