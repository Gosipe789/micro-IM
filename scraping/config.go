package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config struct {
	Mysql Mysql `json:"mysql" yaml:"mysql"`
}

type Mysql struct {
	DriverName     string `json:"driver_name" yaml:"driver_name"`
	DataSourceName string `json:"data_source_name" yaml:"data_source_name"`
}

func (c *Config) getConf() (*Config, error) {
	//应该是 绝对地址
	yamlFile, err := ioutil.ReadFile("./scraping/config.yaml")
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(yamlFile, c)

	if err != nil {
		return nil, err
	}

	return c, nil
}
