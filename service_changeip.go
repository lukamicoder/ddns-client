package main

import (
	"errors"
)

type changeIPService Ddns

func (s *changeIPService) getDomain() string {
	return s.Domain
}

func (s *changeIPService) updateIP() error {
	url := "https://nic.changeip.com/nic/update?u=" + s.Account + "&p=" + s.Password + "&cmd=update&hostname=" + s.Domain

	content, err := GetResponse(url, "", "")
	if err != nil {
		return err
	}

	if content != "200 Successful Update" {
		return errors.New(content)
	}

	return nil
}
