package main

import (
	"encoding/xml"
	"errors"
	"strings"
)

type NameCheapService Ddns

type dictionary struct {
	ErrCount int `xml:"ErrCount"`
}

func (s *NameCheapService) getDomain() string {
	return s.Domain
}

func (s *NameCheapService) updateIP() error {
	pos := strings.Index(s.Domain, ".")
	if pos < 1 {
		return errors.New("Incorrect domain.")
	}

	host := s.Domain[0:pos]
	domain := s.Domain[pos+1 : len(s.Domain)]

	url := "https://dynamicdns.park-your-domain.com/update?domain=" + domain + "&host=" + host + "&password=" + s.Password

	content, err := GetResponse(url, "", "")
	if err != nil {
		return err
	}

	var dict dictionary
	err = xml.Unmarshal([]byte(content), &dict)
	if err != nil {
		return err
	}

	if dict.ErrCount > 0 {
		return errors.New("Unable to update ip address.")
	}

	return nil
}

//<?xml version="1.0"?>
//<interface-response>
// <Command>SETDNSHOST</Command>
// <Language>eng</Language>
// <IP>x.x.x.x</IP>
// <ErrCount>0</ErrCount>
// <ResponseCount>0</ResponseCount>
// <Done>true</Done>
// <debug><![CDATA[]]></debug>
//</interface-response>
