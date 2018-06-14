package config

import (
	"errors"

	"github.com/go-yaml/yaml"
)

// GetConfigFromYAML parses yaml and converts to a go struct
func GetConfigFromYAML(yamlStr []byte) (*Config, error) {
	if yamlStr == nil {
		return nil, errors.New("please provide a yaml string to parse")
	}
	var zuesBaseConfig Config
	err := yaml.Unmarshal(yamlStr, &zuesBaseConfig)
	if err != nil {
		return nil, errors.New("unable to parse yaml. provide a valid yaml string")
	}

	return &zuesBaseConfig, nil
}
