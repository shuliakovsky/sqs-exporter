// config.go
package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type QueueConfig struct {
	URL    string `yaml:"url"`
	Region string `yaml:"region,omitempty"`
}

type Config struct {
	AWSRegion string        `yaml:"aws_region"`
	ListenIP  string        `yaml:"listen_ip,omitempty"`
	Port      int           `yaml:"port,omitempty"`
	Queues    []QueueConfig `yaml:"queues"`
}

func LoadConfig(configFilePath string) (Config, error) {
	configData, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return Config{}, err
	}
	var config Config
	if err := yaml.Unmarshal(configData, &config); err != nil {
		return Config{}, err
	}

	if config.Port == 0 {
		config.Port = 9090
	}
	if config.ListenIP == "" {
		config.ListenIP = "0.0.0.0"
	}
	return config, nil
}
