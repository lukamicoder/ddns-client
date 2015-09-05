package main

import (
	"bitbucket.org/kardianos/service"
	"errors"
	"fmt"
	"github.com/lukamicoder/ini-parser"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	interval = 3600
	services []IDdns
	log service.Logger

	regex = regexp.MustCompile("(?m)[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}")
	logLevel = "info"
	exit = make(chan struct {})
)

var urls = []string{
	"myipinfo.net",
	"myip.dnsomatic.com",
	"icanhazip.com",
	"checkip.dyndns.org",
	"www.myipnumber.com",
	"checkmyip.com",
	"myexternalip.com",
	"www.ipchicken.com",
	"ipecho.net/plain",
	"bot.whatismyipaddress.com",
	"ipv4.ipogre.com",
	"smart-ip.net/myip",
	"checkip.amazonaws.com",
	"www.checkip.org",
}

//Ddns represents a dynamic DNS service
type Ddns struct {
	Name     string
	Domain   string
	Account  string
	Password string
	Token    string
}

//IDdns represents an interface for a dynamic DNS service
type IDdns interface {
	updateIP() error
	getDomain() string
}

func main() {
	var name = "ddns-client"
	var displayName = "DDNS Client"
	var desc = "Dynamic DNS Client."

	rand.Seed(time.Now().UTC().UnixNano())

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
			if logLevel != "info" && logLevel != "error" && logLevel != "nolog" {
				logLevel = "info"
				log.Error("Incorrect loglevel in config file: %v", logLevel)
			}
			interval, err = config.GetInt(name, "interval")
			if err != nil {
				logMessage("error", err.Error())
			}
			continue
		}

		t, err := config.GetString(name, "type")
		if err != nil {
			logMessage("error", err.Error())
			continue
		}
		switch strings.ToLower(t) {
		case "namecheap":
			service := new(nameCheapService)
			service.Name = name
			service.Domain, err = config.GetString(name, "domain")
			if err != nil {
				logMessage("error", "%s - %s", name, err)
				continue
			}
			_, err := net.LookupHost(service.getDomain())
			if err != nil {
				logMessage("error", "%s - %s", name, err)
				continue
			}

			service.Password, err = config.GetString(name, "password")
			if err != nil {
				logMessage("error", "%s - %s", name, err)
				continue
			}
			services = append(services, service)
		case "noip":
		case "no-ip":
			service := new(noIPService)
			service.Name = name
			service.Domain, err = config.GetString(name, "domain")
			if err != nil {
				logMessage("error", "%s - %s", name, err)
				continue
			}
			_, err := net.LookupHost(service.getDomain())
			if err != nil {
				logMessage("error", "%s - %s", name, err)
				continue
			}

			service.Account, err = config.GetString(name, "account")
			if err != nil {
				logMessage("error", "%s - %s", name, err)
				continue
			}
			service.Password, err = config.GetString(name, "password")
			if err != nil {
				logMessage("error", "%s - %s", name, err)
				continue
			}
			services = append(services, service)
		case "changeip":
			service := new(changeIPService)
			service.Name = name
			service.Domain, err = config.GetString(name, "domain")
			if err != nil {
				logMessage("error", "%s - %s", name, err)
				continue
			}
			_, err := net.LookupHost(service.getDomain())
			if err != nil {
				logMessage("error", "%s - %s", name, err)
				continue
			}

			service.Account, err = config.GetString(name, "account")
			if err != nil {
				logMessage("error", "%s - %s", name, err)
				continue
			}
			service.Password, err = config.GetString(name, "password")
			if err != nil {
				logMessage("error", "%s - %s", name, err)
				continue
			}
			services = append(services, service)
		case "duckdns":
			service := new(duckDNSService)
			service.Name = name
			service.Domain, err = config.GetString(name, "domain")
			if err != nil {
				logMessage("error", "%s - %s", name, err)
				continue
			}
			_, err := net.LookupHost(service.getDomain())
			if err != nil {
				logMessage("error", "%s - %s", name, err)
				continue
			}

			service.Token, err = config.GetString(name, "token")
			if err != nil {
				logMessage("error", "%s - %s", name, err)
				continue
			}
			services = append(services, service)
		case "freedns":
			service := new(freeDNSService)
			service.Name = name
			service.Domain, err = config.GetString(name, "domain")
			if err != nil {
				logMessage("error", "%s - %s", name, err)
				continue
			}
			_, err := net.LookupHost(service.getDomain())
			if err != nil {
				logMessage("error", "%s - %s", name, err)
				continue
			}

			service.Token, err = config.GetString(name, "token")
			if err != nil {
				logMessage("error", "%s - %s", name, err)
				continue
			}
			services = append(services, service)
		case "system-ns":
		case "systemns":
			service := new(systemNSService)
			service.Name = name
			service.Domain, err = config.GetString(name, "domain")
			if err != nil {
				logMessage("error", "%s - %s", name, err)
				continue
			}
			_, err := net.LookupHost(service.getDomain())
			if err != nil {
				logMessage("error", "%s - %s", name, err)
				continue
			}
			service.Token, err = config.GetString(name, "token")
			if err != nil {
				logMessage("error", "%s - %s", name, err)
				continue
			}
			services = append(services, service)
		case "ipdns":
			service := new(ipDNSService)
			service.Name = name
			service.Domain, err = config.GetString(name, "domain")
			if err != nil {
				logMessage("error", "%s - %s", name, err)
				continue
			}
			_, err := net.LookupHost(service.getDomain())
			if err != nil {
				logMessage("error", "%s - %s", name, err)
				continue
			}
			service.Password, err = config.GetString(name, "password")
			if err != nil {
				logMessage("error", "%s - %s", name, err)
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
	if logLevel == "nolog" {
		return
	}

	if logLevel == "info" && level == "info" {
		log.Info(format, a...)
		return
	}

	if level == "error" {
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
	logMessage("info", "Stopping service...")
	exit <- struct {}{}
}

func update() {
	currentIP := getExternalIP()
	if currentIP == nil {
		return
	}

	for _, service := range services {
		addr, err := net.LookupHost(service.getDomain())
		if err != nil {
			logMessage("error", "%s - %s", service.getDomain(), err)
			continue
		}
		if len(addr) == 0 || addr[0] == "" {
			logMessage("error", "%s - Unable to get IP address", service.getDomain())
			continue
		}

		registeredIP := addr[0]

		if currentIP.String() == registeredIP {
			logMessage("info", "%s - No update is necessary", service.getDomain())
		} else {
			err := service.updateIP()
			if err == nil {
				logMessage("info", "%s - Successfully updated from %s to %s", service.getDomain(), registeredIP, currentIP.String())
			} else {
				logMessage("error", "%s - %s", service.getDomain(), err)
			}
		}
	}
}

func getExternalIP() net.IP {
	var currentIP net.IP
	for _, i := range rand.Perm(len(urls)) {
		url := urls[i]

		content, err := GetResponse(url, "", "")
		if err != nil {
			logMessage("error", "%s - %s", url, err)
			continue
		}

		ip := regex.FindString(content)

		currentIP = net.ParseIP(ip)

		if currentIP != nil {
			return currentIP
		}
	}

	return currentIP
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
