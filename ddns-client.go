package main

import (
	"github.com/kardianos/service"
	"errors"
	"github.com/lukamicoder/ini-parser"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"
	"log"
	"os"
)

var (
	interval = 3600
	services []DdnsService
	logger service.Logger

	regex = regexp.MustCompile("(?m)[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}")
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

//DdnsService represents an interface for a dynamic DNS service
type DdnsService interface {
	updateIP() error
	getDomain() string
}

type program struct {
	exit chan struct{}
}

func (p *program) Start(s service.Service) error {
	p.exit = make(chan struct{})
	go p.run()
	return nil
}

func (p *program) run() error {
	update()
	ticker := time.NewTicker(time.Duration(interval) * time.Second)

	for {
		select {
		case <-ticker.C:
			update()
		case <-p.exit:
			ticker.Stop()
			return nil
		}
	}
}

func (p *program) Stop(s service.Service) error {
	logger.Info("Stopping service...")
	close(p.exit)
	return nil
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	
	svcConfig := &service.Config{
		Name:       "ddns-client",
		DisplayName: "DDNS Client",
		Description: "Dynamic DNS Client.",
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	
	logger, err = s.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}
	
	err = loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	if len(os.Args) > 1 {
		err := service.Control(s, os.Args[1])
		if err != nil {
			log.Printf("Valid actions: %q\n", service.ControlAction)
			log.Fatal(err)
		}
		
		return
	}
	
	err = s.Run()
	if err != nil {
		logger.Error(err)
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
			interval, err = config.GetInt(name, "interval")
			if err != nil {
				logger.Errorf("%s - %s", name, err)
			}
			continue
		}

		t, err := config.GetString(name, "type")
		if err != nil {
			logger.Error(err)
			continue
		}
		switch strings.ToLower(t) {
		case "namecheap":
			service := new(nameCheapService)
			service.Name = name
			service.Domain, err = config.GetString(name, "domain")
			if err != nil {
				
				logger.Errorf("%s - %s", name, err)
				continue
			}
			_, err := net.LookupHost(service.getDomain())
			if err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}

			service.Password, err = config.GetString(name, "password")
			if err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			services = append(services, service)
		case "noip":
		case "no-ip":
			service := new(noIPService)
			service.Name = name
			service.Domain, err = config.GetString(name, "domain")
			if err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			_, err := net.LookupHost(service.getDomain())
			if err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}

			service.Account, err = config.GetString(name, "account")
			if err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			service.Password, err = config.GetString(name, "password")
			if err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			services = append(services, service)
		case "changeip":
			service := new(changeIPService)
			service.Name = name
			service.Domain, err = config.GetString(name, "domain")
			if err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			_, err := net.LookupHost(service.getDomain())
			if err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}

			service.Account, err = config.GetString(name, "account")
			if err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			service.Password, err = config.GetString(name, "password")
			if err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			services = append(services, service)
		case "duckdns":
			service := new(duckDNSService)
			service.Name = name
			service.Domain, err = config.GetString(name, "domain")
			if err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			_, err := net.LookupHost(service.getDomain())
			if err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}

			service.Token, err = config.GetString(name, "token")
			if err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			services = append(services, service)
		case "freedns":
			service := new(freeDNSService)
			service.Name = name
			service.Domain, err = config.GetString(name, "domain")
			if err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			_, err := net.LookupHost(service.getDomain())
			if err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}

			service.Token, err = config.GetString(name, "token")
			if err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			services = append(services, service)
		case "system-ns":
		case "systemns":
			service := new(systemNSService)
			service.Name = name
			service.Domain, err = config.GetString(name, "domain")
			if err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			_, err := net.LookupHost(service.getDomain())
			if err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			service.Token, err = config.GetString(name, "token")
			if err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			services = append(services, service)
		case "ipdns":
			service := new(ipDNSService)
			service.Name = name
			service.Domain, err = config.GetString(name, "domain")
			if err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			_, err := net.LookupHost(service.getDomain())
			if err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			service.Password, err = config.GetString(name, "password")
			if err != nil {
				logger.Errorf("%s - %s", name, err)
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

func update() {
	currentIP := getExternalIP()
	if currentIP == nil {
		return
	}

	for _, service := range services {
		addr, err := net.LookupHost(service.getDomain())
		if err != nil {
				logger.Errorf("%s - %s", service.getDomain(), err)
			continue
		}
		if len(addr) == 0 || addr[0] == "" {
			logger.Errorf("%s - Unable to get IP address", service.getDomain())
			continue
		}

		registeredIP := addr[0]

		if currentIP.String() == registeredIP {
			logger.Infof("%s - No update is necessary", service.getDomain())
		} else {
			err := service.updateIP()
			if err == nil {
				logger.Infof("%s - Successfully updated from %s to %s", service.getDomain(), registeredIP, currentIP.String())
			} else {
				logger.Errorf("%s - %s", service.getDomain(), err)
			}
		}
	}
}

func getExternalIP() net.IP {
	var currentIP net.IP
	for _, i := range rand.Perm(len(urls)) {
		url := "http://" + urls[i]

		content, err := GetResponse(url, "", "")
		if err != nil {
			logger.Errorf("%s - %s", url, err)
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
