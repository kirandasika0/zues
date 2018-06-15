package kube

// APIServer is the main K8s REST API server endpoint
const APIServer = "http://localhost:8001"

// K8sSession describe all the parameters need to keep a valid connection
// to the Kubernetes server
type K8sSession struct {
	ServerAddress string `json:"server_address"`
	ServerPort    uint16 `json:"server_port"`
	ServerBaseURL string `json:"server_base_url"`
	AccessToken   string `json:"access_token"`
}

// Pod struct represents a K8s pod
type Pod struct {
	Metadata struct {
		Name         string `json:"name"`
		GenerateName string `json:"generateName"`
		Namespace    string `json:"namespace"`
		SelfLink     string `json:"selfLink"`
	} `json:"metadata"`
}

// Service represents a K8s service that is returned by the API.
type Service struct {
	Metadata struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		SelfLink  string `json:"selfLink"`
	} `json:"metadata"`
}

// Services API Endpoint object parse all services in the cluster
type Services struct {
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`
	MetaData   struct {
		SelfLink        string `json:"selfLink"`
		ResourceVersion int64  `json:"resourceVersion"`
	} `json:"metadata"`
	Items []Service `json:"items"`
}
