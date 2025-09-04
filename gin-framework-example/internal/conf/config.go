package conf

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

var Conf *Config

type Config struct {
	App AppConfig `yaml:"app"`
}

type AppConfig struct {
	Name string `yaml:"name"`
	Port int    `yaml:"port"`
}

func Init() error {
	data, err := ioutil.ReadFile("internal/conf/config.yaml")
	if err != nil {
		return err
	}

	Conf = &Config{}
	return yaml.Unmarshal(data, Conf)
}
