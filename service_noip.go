package main

import (
	"errors"
	"strings"
)

type NoIPService Ddns

func (s *NoIPService) GetDomain() string {
	return s.Domain
}

func (s *NoIPService) UpdateIP() error {
	content, err := GetContent("https://dynupdate.no-ip.com/nic/update?hostname="+s.Domain, s.Account, s.Password)
	if err != nil {
		return err
	}

	body := string(content)
	if !strings.HasPrefix(body, "nochg") && !strings.HasPrefix(body, "good") {
		return errors.New(body)
	}

	return nil
}
