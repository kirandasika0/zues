package kube

import (
	"flag"
	"path/filepath"
	"strings"
	"zues/util"

	"k8s.io/apimachinery/pkg/watch"

	"github.com/kataras/golog"

	"os"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// Sessionv1 is a struct to represent the global K8s session
type Sessionv1 struct {
	clientSet  *kubernetes.Clientset
	apiCalls   uint64
	hasStarted bool
}

var (
	// Session is  used as a global variables throughout the program
	Session *Sessionv1
)

// New Create a new kuberentes session
func New() *Sessionv1 {
	var kubeconfig *string
	if os.Getenv("IN_CLUSTER") == "true" {
		// TODO
	} else {
		if home := homeDir(); home != "" {
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		} else {
			kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
		}
		flag.Parse()
	}
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())

	}
	// Create a new session
	var s Sessionv1
	s.apiCalls = 0
	s.clientSet = clientset
	s.hasStarted = false

	return &s
}

// CreatePod create a pod
func (s *Sessionv1) CreatePod(serviceName, namespace string, labels map[string]string, container apiv1.Container) (*apiv1.Pod, error) {
	podName := strings.ToLower(serviceName + "-" + util.RandomString(5) + "-" + util.RandomString(5))
	pod, err := s.clientSet.CoreV1().Pods(namespace).Create(&apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: namespace,
		},
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{
				container,
			},
			ImagePullSecrets: []apiv1.LocalObjectReference{
				{
					// TODO: remove this and make this a environment variable
					Name: "secqat",
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return pod, nil
}

// GetPod gets the specific pod in a namespace
func (s *Sessionv1) GetPod(podName, namespace string) (*apiv1.Pod, error) {
	if len(namespace) == 0 {
		namespace = "default"
	}
	pod, err := s.clientSet.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	s.apiCalls++
	return pod, nil
}

// DeletePod deletes the pod in a specific namespace
func (s *Sessionv1) DeletePod(podName, namespace string) error {
	if len(namespace) == 0 {
		namespace = "default"
	}
	err := s.clientSet.CoreV1().Pods(namespace).Delete(podName, &metav1.DeleteOptions{})
	s.apiCalls++
	if err != nil {
		return err
	}
	return nil
}

// ListPods lists all the pods in a given namespace
func (s *Sessionv1) ListPods(namespace string) (*apiv1.PodList, error) {
	if len(namespace) == 0 {
		namespace = "default"
	}
	podList, err := s.clientSet.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	s.apiCalls++
	if err != nil {
		return nil, err
	}

	return podList, err
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE")
}

// CreateContainer creates a K8s container with the given parameters
func CreateContainer(name, image string, port int32) apiv1.Container {
	return apiv1.Container{
		Name:            name,
		Image:           image,
		ImagePullPolicy: apiv1.PullAlways,
		Ports: []apiv1.ContainerPort{
			{
				Name:          "http", // Currently only supporting HTTP servers
				Protocol:      apiv1.ProtocolTCP,
				ContainerPort: port,
			},
		},
	}
}

// WatchPodEvents watches events on Pods
func (s *Sessionv1) WatchPodEvents() {
	golog.Println("Watching Pod Events...")
	watcher, err := s.clientSet.CoreV1().Pods("default").Watch(metav1.ListOptions{})
	if err != nil {
		golog.Errorf("Error occured: %s", err.Error())
	}
	for event := range watcher.ResultChan() {
		pod, ok := event.Object.(*apiv1.Pod)
		if !ok {
			golog.Error("Failed to convert to a type of apiv1.Pod")
			continue
		}
		switch event.Type {
		case watch.Added:
			golog.Infof("Pod %s added", pod.ObjectMeta.Name)
		case watch.Modified:
			golog.Infof("Pod %s modified", pod.ObjectMeta.Name)
			for _, c := range pod.Status.ContainerStatuses {
				golog.Info(c.State.Waiting, c.State.Running, c.State.Terminated)
			}
		case watch.Error:
			golog.Errorf("Error in pod %s", pod.ObjectMeta.Name)
		case watch.Deleted:
			golog.Infof("Pod %s deleted", pod.ObjectMeta.Name)
		}
	}

}
