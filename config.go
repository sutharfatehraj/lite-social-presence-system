package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Config struct to hold configuration values
type Config struct {
	RestAPIServerAddress string `yaml:"rest_api_server_address"`
	GRPCNetwork          string `yaml:"grpc_network"`
	GRPCServerAddress    string `yaml:"grpc_server_address"`
	MongoURI             string `yaml:"mongo_uri"`
}

// LoadConfig function to read from the YAML file
func LoadConfig(filePath string) (*Config, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
