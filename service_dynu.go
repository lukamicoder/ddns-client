package main

import (
	"errors"
	"strings"
)

type dynuService Ddns

func (s *dynuService) getDomain() string {
	return s.Domain
}

func (s *dynuService) updateIP() error {
	content, err := GetResponse("https://api.dynu.com/nic/update?hostname=" + s.Domain, s.Account, s.Password)
	if err != nil {
		return err
	}

	if !strings.HasPrefix(content, "nochg") && !strings.HasPrefix(content, "good") {
		return errors.New(content)
	}

	return nil
}
