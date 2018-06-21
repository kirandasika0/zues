package kube

import (
	"testing"

	appv1 "k8s.io/api/core/v1"
)

var testContainer = appv1.Container{
	Name:    "ubuntu",
	Image:   "ubuntu:trusty",
	Command: []string{"echo"},
	Args:    []string{"Hello, world"},
}

var s *Sessionv1

func TestNew(t *testing.T) {
	newSession := New()
	if newSession.clientSet == nil {
		t.Error("Could not create the client set.")
	}
	t.Log("Passed :)")
	s = newSession
}

func TestCreatePod(t *testing.T) {
	pod, err := s.CreatePod("pod-example", "default", nil, testContainer)
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}

	s.DeletePod(pod.ObjectMeta.Name, "default")
}

func TestDeletePod(t *testing.T) {
	pod, _ := s.CreatePod("pod-example", "default", nil, testContainer)

	err := s.DeletePod(pod.ObjectMeta.Name, "")
	if err != nil {
		t.Fail()
	}
}

func TestListPods(t *testing.T) {
	pods, err := s.ListPods("default")
	if pods.Items == nil || err != nil {
		t.Fail()
	}
}
