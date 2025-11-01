package util

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

var Conf *Config

type Config struct {
	App   AppConfig   `yaml:"app"`
	DB    DBConfig    `yaml:"db"`
	Redis RedisConfig `yaml:"redis"`
}

type AppConfig struct {
	Name string `yaml:"name"`
	Port int    `yaml:"port"`
}

type DBConfig struct {
	Driver   string `yaml:"driver"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	Charset  string `yaml:"charset"`
	Prefix   string `yaml:"prefix"`
	SSLMode  string `yaml:"sslmode"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

func Init(env string) error {
	path := fmt.Sprintf("conf/config.%s.yaml", env)

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	Conf = &Config{}
	return yaml.Unmarshal(data, Conf)
}
