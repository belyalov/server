package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Config is top level configuration for all features
type Config struct {
	UDP map[string]interface{}
}

// LoadConfigFromFile reads and parses YAML configuration from file
func LoadConfigFromFile(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	cfg := &Config{}
	err = yaml.UnmarshalStrict(data, cfg)

	return cfg, err
}
