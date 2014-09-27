package main

import (
	"errors"
	"strings"
)

type ChangeIPService Ddns

func (s *ChangeIPService) GetDomain() string {
	return s.Domain
}

func (s *ChangeIPService) UpdateIP() error {
	url := []string{"https://nic.changeip.com/nic/update?u=", s.Account, "&p=", s.Password, "&cmd=update&hostname=", s.Domain}

	content, err := GetContent(strings.Join(url, ""), "", "")
	if err != nil {
		return err
	}

	if string(content) != "200 Successful Update" {
		return errors.New(string(content))
	}

	return nil
}
