package main

import (
	"errors"
	"strings"
)

type ddnssdeService Ddns

func (s *ddnssdeService) getDomain() string {
	return s.Domain
}

func (s *ddnssdeService) updateIP() error {
	content, err := GetResponse("http://ddnss.de/upd.php?user=" + s.UserName + "&pwd=" + s.Password + "&host=" + s.Domain, "", "")
	if err != nil {
		return err
	}

	if strings.HasPrefix(content, "Error") {
		return errors.New(content)
	}

	return nil
}
