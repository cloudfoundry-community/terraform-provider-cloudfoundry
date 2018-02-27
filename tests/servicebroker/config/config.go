package config

import (
	"os"

	"github.com/cloudfoundry-incubator/candiedyaml"
)

type Config struct {
	BrokerConfiguration BrokerConfiguration `yaml:"basic_auth_service_broker"`
}

type BrokerConfiguration struct {
	RouteServiceURL string `yaml:"route_service_url"`
	BrokerUserName  string `yaml:"broker_username"`
	BrokerPassword  string `yaml:"broker_password"`
}

func ParseConfig(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}

	var decodedConfig Config
	if err := candiedyaml.NewDecoder(file).Decode(&decodedConfig); err != nil {
		return Config{}, err
	}
	return decodedConfig, nil
}
