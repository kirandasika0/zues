package kube

import (
	"encoding/json"
	"errors"

	"fmt"
	"os"

	"github.com/tidwall/gjson"
)

// New creates a K8s session and returns a pointer to it
func New() *K8sSession {
	var session K8sSession
	// Check environment variables to see if we have to run in-cluster
	// configuration
	runInClusterConfig := os.Getenv("ZUES_IN_CLUSTER_CONFIG")
	if len(runInClusterConfig) > 1 {
		// TODO: In-cluster configuration

	} else {
		// Run normal configuration
		// Uses localhost k8s proxy
		session.ServerAddress = "localhost"
		session.ServerPort = 8001
		session.ServerBaseURL = fmt.Sprintf("%s:%d", session.ServerAddress, session.ServerPort)
	}

	return &session
}

// CreateNewPodWithNamespace create a pod by calling the K8s API.
func (s *K8sSession) CreateNewPodWithNamespace(namespace string, podName string) Pod {
	// TODO send a post request to K8s api to create a pod

	return Pod{}
}

// GetPodsFromJSON gets pod info from json string
func GetPodsFromJSON(jsonStr []byte) ([]Pod, error) {
	if len(jsonStr) < 1 {
		return nil, errors.New("please provide json string to decode")
	}
	decodedItemsStr := gjson.Get(string(jsonStr), "items")
	var pods []Pod
	for _, item := range decodedItemsStr.Array() {
		var newPod Pod
		_ = json.Unmarshal([]byte(item.String()), &newPod)
		pods = append(pods, newPod)
	}

	return pods, nil
}

// GetServicesFromAPIServer gets all the services from the K8s API Server and parses them to a kube.Services struct
func GetServicesFromAPIServer(jsonStr []byte) (*Services, error) {
	if len(jsonStr) < 1 {
		return nil, errors.New("please provide a string to decode")
	}

	var services Services
	services.Kind = gjson.Get(string(jsonStr), "kind").String()
	services.APIVersion = gjson.Get(string(jsonStr), "apiVersion").String()

	json.Unmarshal([]byte(gjson.Get(string(jsonStr), "metadata").String()), &services.MetaData)

	for _, item := range gjson.Get(string(jsonStr), "items").Array() {
		var s Service
		_ = json.Unmarshal([]byte(item.String()), &s)
		services.Items = append(services.Items, s)
	}

	return &services, nil
}
