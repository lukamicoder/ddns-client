package main

import (
	"errors"
)

type ChangeIPService Ddns

func (s *ChangeIPService) getDomain() string {
	return s.Domain
}

func (s *ChangeIPService) updateIP() error {
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
