package main

import (
	"errors"
)

type ChangeIPService Ddns

func (s *ChangeIPService) GetDomain() string {
	return s.Domain
}

func (s *ChangeIPService) UpdateIP() error {
	url := "https://nic.changeip.com/nic/update?u=" + s.Account + "&p=" + s.Password + "&cmd=update&hostname=" + s.Domain

	content, err := GetContent(url, "", "")
	if err != nil {
		return err
	}

	if content != "200 Successful Update" {
		return errors.New(content)
	}

	return nil
}
