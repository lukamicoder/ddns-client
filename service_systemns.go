package main

import (
	"errors"
	"strconv"
	"strings"
)

type SystemNSService Ddns

func (s *SystemNSService) GetDomain() string {
	return s.Domain
}

func (s *SystemNSService) UpdateIP() error {
	url := "http://system-ns.com/api?type=dynamic&domain=" + s.Domain + "&command=set&token=" + s.Token

	content, err := GetContent(url, "", "")
	if err != nil {
		return err
	}

	pos := strings.Index(content, ":")
	if pos < 1 {
		return errors.New(content)
	}

	code, err := strconv.Atoi(content[pos+1 : pos+2])
	if err != nil {
		return err
	}

	switch code {
	case 0:
		return nil
	case 1:
		return errors.New("Data invalid")
	case 2:
		return errors.New("Token invalid")
	case 3:
		return errors.New("Domain invalid")
	case 4:
		return errors.New("Auth invalid")
	case 5:
		return errors.New("Wrong ip format")
	case 99:
		return errors.New("Another problem")
	}

	return nil
}
