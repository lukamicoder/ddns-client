package main

import (
	"errors"
	"github.com/kardianos/service"
	"github.com/lukamicoder/ini-parser"
	"log"
	"math/rand"
	"net"
	"os"
	"regexp"
	"strings"
	"time"
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

type program struct {
	exit chan struct {}
}

func (p *program) Start(s service.Service) error {
	p.exit = make(chan struct {})
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
		Name:        "ddns-client",
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

	if err := config.LoadFile("./config.ini"); err != nil {
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
			if interval, err = config.GetInt(name, "interval"); err != nil {
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
			if service.Domain, err = config.GetString(name, "domain"); err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			if _, err := net.LookupHost(service.getDomain()); err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			if service.Password, err = config.GetString(name, "password"); err != nil {
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
			service.UserName, err = config.GetString(name, "username")
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
			if service.Domain, err = config.GetString(name, "domain"); err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			if _, err := net.LookupHost(service.getDomain()); err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			if service.UserName, err = config.GetString(name, "username"); err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			if service.Password, err = config.GetString(name, "password"); err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			services = append(services, service)
		case "duckdns":
			service := new(duckDNSService)
			service.Name = name
			if service.Domain, err = config.GetString(name, "domain"); err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			if _, err := net.LookupHost(service.getDomain()); err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			if service.Token, err = config.GetString(name, "token"); err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			services = append(services, service)
		case "freedns":
			service := new(freeDNSService)
			service.Name = name
			if service.Domain, err = config.GetString(name, "domain"); err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			if _, err := net.LookupHost(service.getDomain()); err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			if service.Token, err = config.GetString(name, "token"); err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			services = append(services, service)
		case "system-ns":
		case "systemns":
			service := new(systemNSService)
			service.Name = name
			if service.Domain, err = config.GetString(name, "domain"); err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			if _, err := net.LookupHost(service.getDomain()); err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			if service.Token, err = config.GetString(name, "token"); err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			services = append(services, service)
		case "ipdns":
			service := new(ipDNSService)
			service.Name = name
			if service.Domain, err = config.GetString(name, "domain"); err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			if _, err := net.LookupHost(service.getDomain()); err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			if service.UserName, err = config.GetString(name, "username"); err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			if service.Password, err = config.GetString(name, "password"); err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			services = append(services, service)
		case "dynu":
			service := new(dynuService)
			service.Name = name
			if service.Domain, err = config.GetString(name, "domain"); err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			if _, err := net.LookupHost(service.getDomain()); err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			if service.UserName, err = config.GetString(name, "username"); err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			if service.Password, err = config.GetString(name, "password"); err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			services = append(services, service)
		case "ydns":
			service := new(yDNSService)
			service.Name = name
			if service.Domain, err = config.GetString(name, "domain"); err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			if _, err := net.LookupHost(service.getDomain()); err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			if service.UserName, err = config.GetString(name, "username"); err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			if service.Password, err = config.GetString(name, "password"); err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			services = append(services, service)
		case "ddnss.de":
		case "ddnssde":
			service := new(ddnssdeService)
			service.Name = name
			if service.Domain, err = config.GetString(name, "domain"); err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			if _, err := net.LookupHost(service.getDomain()); err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			if service.UserName, err = config.GetString(name, "username"); err != nil {
				logger.Errorf("%s - %s", name, err)
				continue
			}
			if service.Password, err = config.GetString(name, "password"); err != nil {
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
	var currentIP net.IP
	if currentIP = getExternalIP(); currentIP == nil {
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

		if currentIP = net.ParseIP(ip); currentIP != nil {
			return currentIP
		}
	}

	return currentIP
}
