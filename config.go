package main

import (
	"encoding/base64"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
)

type HostConfig struct {
	Username        string `yaml:"username"`
	Password        string `yaml:"password"`
	Hostname        string
	Token           string
	Metrics         string
	Initialized     bool
	Reachable       bool
	Retries         uint32
	SystemEndpoint  string
	ThermalEndpoint string
	PowerEndpoint   string
}

type OmeConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Hostname string `yaml:"hostname"`
	host     HostConfig
}

type RootConfig struct {
	Hostname string                 `yaml:"hostname"`
	Address  string                 `yaml:"address"`
	Port     uint32                 `yaml:"port"`
	Metrics  []string               `yaml:"metrics"`
	Timeout  uint32                 `yaml:"timeout"`
	Retries  uint32                 `yaml:"retries"`
	Hosts    map[string]*HostConfig `yaml:"hosts"`
	Ome      *OmeConfig             `yaml:"ome"`
}

var config RootConfig
var collectSystem bool = false
var collectSensors bool = false
var collectSEL bool = false
var collectPower bool = false

func validateMetrics(name string) bool {
	switch name {
	case "system":
		collectSystem = true
		return true
	case "sensors":
		collectSensors = true
		return true
	case "power":
		collectPower = true
		return true
	case "sel":
		collectSEL = true
		return true
	}
	return false
}

func parseError(s0, s1 string) {
	log.Printf("Error parsing configuration file: %s: %s", s0, s1)
	os.Exit(1)
}

func readConfigFile(fileName string) {
	yamlFile, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Printf("Error reading configuration file: %s\n", err)
		os.Exit(1)
	}

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Printf("Error parsing configuration file: %s\n", err)
		os.Exit(1)
	}

	if len(config.Address) == 0 {
		config.Address = "0.0.0.0"
	}

	if config.Port == 0 {
		config.Port = 9348
	}

	if config.Timeout == 0 {
		config.Timeout = 10
	}

	if config.Retries == 0 {
		config.Retries = 1
	}

	if len(config.Metrics) == 0 {
		parseError("missing section", "metrics")
	}

	if len(config.Hosts) == 0 {
		parseError("missing section", "hosts")
	}

	for _, v := range config.Metrics {
		if !validateMetrics(v) {
			parseError("invalid metrics name", v)
		}
	}

	for k, v := range config.Hosts {
		if len(v.Username) == 0 {
			parseError("missing username for host", k)
		}
		if len(v.Password) == 0 {
			parseError("missing password for host", k)
		}

		data := []byte(v.Username + ":" + v.Password)
		v.Token = base64.StdEncoding.EncodeToString(data)
		v.Hostname = k
		v.Initialized = false
		v.Retries = 0
	}

	if config.Ome != nil {
		if len(config.Ome.Username) == 0 {
			parseError("OME", "missing username")
		}
		if len(config.Ome.Password) == 0 {
			parseError("OME", "missing password")
		}
		if len(config.Ome.Hostname) == 0 {
			parseError("OME", "missing hostname")
		}
		config.Ome.host.Token = base64.StdEncoding.EncodeToString(
			[]byte(config.Ome.Username + ":" + config.Ome.Password))
		config.Ome.host.Hostname = config.Ome.Hostname
	}
}
