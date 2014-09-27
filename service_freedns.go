package main

import (
	"errors"
	"strings"
)

type FreeDNSService Ddns

func (s *FreeDNSService) GetDomain() string {
	return s.Domain
}

func (s *FreeDNSService) UpdateIP() error {
	url := []string{"http://freedns.afraid.org/dynamic/update.php?", s.Token}

	content, err := GetContent(strings.Join(url, ""), "", "")
	if err != nil {
		return err
	}

	if string(content)[0:4] == "ERROR" {
		return errors.New(string(content))
	}

	return nil
}
