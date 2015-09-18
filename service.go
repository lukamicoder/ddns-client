package main

import (
	"io/ioutil"
	"net/http"
)

//Ddns represents a dynamic DNS service
type Ddns struct {
	Name     string
	Domain   string
	Account  string
	Password string
	Token    string
}

//DdnsService represents an interface for a dynamic DNS service
type DdnsService interface {
	updateIP() error
	getDomain() string
}

//GetResponse returns the content at the url address
func GetResponse(url string, login string, password string) (string, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	if login != "" && password != "" {
		request.SetBasicAuth(login, password)
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()
	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(content), nil
}
