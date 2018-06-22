package config

import (
	"errors"
	"fmt"

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

// MatchJobIDWithPod returns the JobID given a podName
func MatchJobIDWithPod(podName string) (string, error) {
	for k, v := range JobPodsMap {
		if v.ObjectMeta.Name == podName {
			return k, nil
		}
	}
	return "", fmt.Errorf("Could not find JobID for given Pod %s", podName)
}
