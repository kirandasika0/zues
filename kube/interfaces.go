package kube

import (
	"encoding/json"
	"errors"
	"zues/util"

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
		// Make sure it is running on port 8001
		// COMMAND: kubectl proxy --port=8001
		session.ServerAddress = "localhost"
		session.ServerPort = 8001
		session.ServerBaseURL = fmt.Sprintf("http://%s:%d", session.ServerAddress, session.ServerPort)
		session.DefaultHeaders = map[string]string{
			"Content-Type": "application/json",
		}
	}

	return &session
}

// Execute runs the current K8s session in a background thread
func (s *K8sSession) Execute() error {
	// TODO : run a test ping to the API server

	return nil
}

// Kill destroys the current K8s session that is running in the background
func (s *K8sSession) Kill() error {
	// TODO : add more functionality

	return nil
}

// CreateNewPodWithNamespace create a pod by calling the K8s API.
func (s *K8sSession) CreateNewPodWithNamespace(namespace string, podName string) Pod {
	// TODO send a post request to K8s api to create a pod

	return Pod{}
}

// GetPodsFromAPIServer gets pod info from json string
func (s *K8sSession) GetPodsFromAPIServer(jsonStr []byte) ([]Pod, error) {
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
func (s *K8sSession) GetServicesFromAPIServer(jsonStr []byte) (*Services, error) {
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

// CreatePodWithNamespace creates a pod on a cluster in the given namespace
func (s *K8sSession) CreatePodWithNamespace(podData []byte, namespace string) (Pod, error) {
	// Steps to create a pod on a cluster
	// 1. Create a unique name for the pod
	// 2. Check if the namespace is specified in the request
	// 3. Default to the namespace specified in the zues setup config
	// 4. Save metadata returned from the K8s API in a base64 style

	// Acccess the K8s API server to create a pod with the given spec
	req, err := util.CreateHTTPRequest("POST",
		s.ServerBaseURL+"/api/v1/namespaces/sprintt-qa/pods",
		s.DefaultHeaders, podData)

	// Define our new pod
	var newPod Pod
	if err != nil {
		return newPod, err
	}

	_, resp, err := util.GetHTTPResponse(req)
	if err != nil {
		return newPod, err
	}

	// Unmarshal the JSON returned by the K8s API
	json.Unmarshal(resp, &newPod)

	return newPod, nil
}

// DeletePodWithNamespace deletes a certain pod from the cluster in a given namespace
func (s *K8sSession) DeletePodWithNamespace(podName string, namespace string, uidIn string) (Pod, error) {
	// This is a temporary pod deleteion spec
	type preconditionsT struct {
		uid string
	}
	type tempDeletePodSpec struct {
		apiVersion         string
		gracePeriodSeconds int
		kind               string
		preconditions      preconditionsT
	}

	var pod = tempDeletePodSpec{
		apiVersion:         "v1",
		gracePeriodSeconds: 10,
		kind:               "Pod",
		preconditions: preconditionsT{
			uid: uidIn,
		},
	}
	deleteData, err := json.Marshal(pod)
	if err != nil {
		return Pod{}, err
	}
	req, err := util.CreateHTTPRequest("DELETE",
		s.ServerBaseURL+"/api/v1/namespaces/"+namespace+"/pods/"+podName,
		s.DefaultHeaders, deleteData)
	var deletedPod Pod
	if err != nil {
		return deletedPod, err
	}
	_, resp, err := util.GetHTTPResponse(req)
	if err != nil {
		return deletedPod, err
	}

	json.Unmarshal(resp, &deletedPod)
	return deletedPod, nil
}
