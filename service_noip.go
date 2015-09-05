package main

import (
	"errors"
	"strings"
)

type noIPService Ddns

func (s *noIPService) getDomain() string {
	return s.Domain
}

func (s *noIPService) updateIP() error {
	content, err := GetResponse("https://dynupdate.no-ip.com/nic/update?hostname="+s.Domain, s.Account, s.Password)
	if err != nil {
		return err
	}

	if !strings.HasPrefix(content, "nochg") && !strings.HasPrefix(content, "good") {
		return errors.New(content)
	}

	return nil
}
