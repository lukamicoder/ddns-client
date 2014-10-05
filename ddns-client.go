package main

import (
	"bitbucket.org/kardianos/service"
	"errors"
	"fmt"
	"io/ioutil"
	"lukamicoder/ini-parser"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	interval int = 3600
	services []IDdns
	log      service.Logger
	logLevel string = "info"
	exit            = make(chan struct{})
)

const (
	NOLOG = "nolog"
	INFO  = "info"
	ERROR = "error"
)

type Ddns struct {
	Name     string
	Domain   string
	Account  string
	Password string
	Token    string
}

type IDdns interface {
	UpdateIP() error
	GetDomain() string
}

func main() {
	var name = "ddns-client"
	var displayName = "DDNS Client"
	var desc = "Dynamic DNS Client."

	var s, err = service.NewService(name, displayName, desc)
	if err != nil {
		fmt.Printf("Unable to start: %s\n", err)
		return
	}
	log = s

	if len(os.Args) > 1 {
		var err error
		verb := os.Args[1]
		switch verb {
		case "install":
			err = s.Install()
			if err != nil {
				fmt.Printf("Failed to install: %s\n", err)
			} else {
				fmt.Printf("Service installed.\n")
			}
		case "remove":
			err = s.Remove()
			if err != nil {
				fmt.Printf("Failed to remove: %s\n", err)
			} else {
				fmt.Printf("Service removed.\n")
			}
		case "start":
			err = s.Start()
			if err != nil {
				fmt.Printf("Failed to start: %s\n", err)
			} else {
				fmt.Printf("Service started.\n")
			}
		case "stop":
			err = s.Stop()
			if err != nil {
				fmt.Printf("Failed to stop: %s\n", err)
			} else {
				fmt.Printf("Service stopped.\n")
			}
		}

		return
	}

	err = s.Run(func() error {
		err = loadConfig()
		if err != nil {
			return err
		}

		go runTicker()
		return nil
	}, func() error {
		stopTicker()
		return nil
	})
	if err != nil {
		s.Error(err.Error())
	}
}

func loadConfig() error {
	var config iniparser.Config

	err := config.LoadFile("./config.ini")
	if err != nil {
		return err
	}

	sections := config.GetSections()
	if len(sections) < 2 {
		return errors.New("No services found in config file.\n")
	}

	for _, section := range sections {
		var err error
		var name = section.Name

		if name == "settings" {
			logLevel, _ = config.GetString(name, "loglevel")
			if logLevel != INFO && logLevel != ERROR && logLevel != NOLOG {
				logLevel = INFO
				log.Error("Incorrect loglevel in config file: %v", logLevel)
			}
			interval, err = config.GetInt(name, "interval")
			if err != nil {
				logMessage(ERROR, err.Error())
			}
			continue
		}

		t, err := config.GetString(name, "type")
		if err != nil {
			logMessage(ERROR, err.Error())
			continue
		}
		switch strings.ToLower(t) {
		case "namecheap":
			service := new(NameCheapService)
			service.Name = name
			service.Domain, err = config.GetString(name, "domain")
			if err != nil {
				logMessage(ERROR, "%s - %s", name, err)
				continue
			}
			_, err := net.LookupHost(service.GetDomain())
			if err != nil {
				logMessage(ERROR, "%s - %s", name, err)
				continue
			}

			service.Password, err = config.GetString(name, "password")
			if err != nil {
				logMessage(ERROR, "%s - %s", name, err)
				continue
			}
			services = append(services, service)
		case "noip":
		case "no-ip":
			service := new(NoIPService)
			service.Name = name
			service.Domain, err = config.GetString(name, "domain")
			if err != nil {
				logMessage(ERROR, "%s - %s", name, err)
				continue
			}
			_, err := net.LookupHost(service.GetDomain())
			if err != nil {
				logMessage(ERROR, "%s - %s", name, err)
				continue
			}

			service.Account, err = config.GetString(name, "account")
			if err != nil {
				logMessage(ERROR, "%s - %s", name, err)
				continue
			}
			service.Password, err = config.GetString(name, "password")
			if err != nil {
				logMessage(ERROR, "%s - %s", name, err)
				continue
			}
			services = append(services, service)
		case "changeip":
			service := new(ChangeIPService)
			service.Name = name
			service.Domain, err = config.GetString(name, "domain")
			if err != nil {
				logMessage(ERROR, "%s - %s", name, err)
				continue
			}
			_, err := net.LookupHost(service.GetDomain())
			if err != nil {
				logMessage(ERROR, "%s - %s", name, err)
				continue
			}

			service.Account, err = config.GetString(name, "account")
			if err != nil {
				logMessage(ERROR, "%s - %s", name, err)
				continue
			}
			service.Password, err = config.GetString(name, "password")
			if err != nil {
				logMessage(ERROR, "%s - %s", name, err)
				continue
			}
			services = append(services, service)
		case "duckdns":
			service := new(DuckDNSService)
			service.Name = name
			service.Domain, err = config.GetString(name, "domain")
			if err != nil {
				logMessage(ERROR, "%s - %s", name, err)
				continue
			}
			_, err := net.LookupHost(service.GetDomain())
			if err != nil {
				logMessage(ERROR, "%s - %s", name, err)
				continue
			}

			service.Token, err = config.GetString(name, "token")
			if err != nil {
				logMessage(ERROR, "%s - %s", name, err)
				continue
			}
			services = append(services, service)
		case "freedns":
			service := new(FreeDNSService)
			service.Name = name
			service.Domain, err = config.GetString(name, "domain")
			if err != nil {
				logMessage(ERROR, "%s - %s", name, err)
				continue
			}
			_, err := net.LookupHost(service.GetDomain())
			if err != nil {
				logMessage(ERROR, "%s - %s", name, err)
				continue
			}

			service.Token, err = config.GetString(name, "token")
			if err != nil {
				logMessage(ERROR, "%s - %s", name, err)
				continue
			}
			services = append(services, service)
		}
	}

	if len(services) < 1 {
		return errors.New("No valid services found in config file.")
	}

	return nil
}

func logMessage(level string, format string, a ...interface{}) {
	if logLevel == NOLOG {
		return
	}

	if logLevel == INFO && level == INFO {
		log.Info(format, a...)
		return
	}

	if level == ERROR {
		log.Error(format, a...)
	}
}

func runTicker() {
	update()
	ticker := time.NewTicker(time.Duration(interval) * time.Second)

	for {
		select {
		case <-ticker.C:
			update()
		case <-exit:
			ticker.Stop()
			return
		}
	}
}

func stopTicker() {
	logMessage(INFO, "Stopping service...")
	exit <- struct{}{}
}

func update() {
	currentIp := getExternalIP()
	if currentIp == "" {
		return
	}

	for _, service := range services {
		addr, err := net.LookupHost(service.GetDomain())
		if err != nil {
			logMessage(ERROR, "%s - %s", service.GetDomain(), err)
			continue
		}
		if len(addr) == 0 || addr[0] == "" {
			logMessage(ERROR, "%s - Unable to get IP address", service.GetDomain())
			continue
		}

		registeredIp := addr[0]

		if currentIp == registeredIp {
			logMessage(INFO, "%s - No update is necessary", service.GetDomain())
		} else {
			err := service.UpdateIP()
			if err == nil {
				logMessage(INFO, "%s - Successfully updated from %s to %s", service.GetDomain(), registeredIp, currentIp)
			} else {
				logMessage(ERROR, "%s - %s", service.GetDomain(), err)
			}
		}
	}
}

func getExternalIP() string {
	url := "http://myipinfo.net//"
	content, err := GetContent(url, "", "")
	if err == nil {
		html := string(content)
		startPos := strings.Index(html, "<h2>")
		endPos := strings.Index(html, "</h2>")
		if startPos > 0 && endPos > startPos {
			sip := html[startPos+4 : endPos]
			ip := net.ParseIP(sip)

			return ip.String()
		} else {
			logMessage(ERROR, "%s - Parsing failed", url)
		}
	}
	logMessage(ERROR, "%s - %s", url, err)

	url = "http://myip.dnsomatic.com/"
	content, err = GetContent(url, "", "")
	if err == nil {
		html := string(content)
		ip := net.ParseIP(strings.TrimSpace(html))

		return ip.String()
	}
	logMessage(ERROR, "%s - %s", url, err)

	url = "http://icanhazip.com/"
	content, err = GetContent(url, "", "")
	if err == nil {
		html := string(content)
		ip := net.ParseIP(strings.TrimSpace(html))

		return ip.String()
	}
	logMessage(ERROR, "%s - %s", url, err)

	url = "http://checkip.dyndns.org/"
	content, err = GetContent(url, "", "")
	if err == nil {
		html := string(content)
		startPos := strings.Index(html, ": ")
		endPos := strings.Index(html, "</body>")
		if startPos > 0 && endPos > startPos {
			sip := html[startPos+2 : endPos]
			ip := net.ParseIP(sip)

			return ip.String()
		} else {
			logMessage(ERROR, "%s - Parsing failed", url)
		}
	}
	logMessage(ERROR, "%s - %s", url, err)

	return ""
}

func GetContent(url string, login string, password string) ([]byte, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if login != "" && password != "" {
		request.SetBasicAuth(login, password)
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return content, nil
}
