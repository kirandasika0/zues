package server

import (
	"errors"
	"zues/kube"

	"github.com/kataras/golog"
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	recover2 "github.com/kataras/iris/middleware/recover"
)

var (
	ZuesServer *Server
)

// Server holds all the necessary information for the zues HTTP API to function
type Server struct {
	Application   *iris.Application
	Port          string
	Configuration iris.Configuration
	K8sSession    *kube.K8sSession
}

// New creates a new instance of the zues server
func New(serverConfig *iris.Configuration, serverPort string) *Server {
	var s Server
	if serverConfig == nil {
		s.Configuration = getDefaultIrisConfiguration()
	}

	// Start up a new Iris Application
	s.Application = iris.New()
	s.Application.Logger().SetLevel("debug")
	s.Application.Use(recover2.New())
	s.Application.Use(logger.New())

	// Register all the routes to the server
	registerRoutes(&s)

	if len(serverPort) < 1 {
		// Set default port to localhost 8284
		s.Port = ":8284"
	} else {
		s.Port = serverPort
	}

	return &s
}

// Start starts the zues server
func (s *Server) Start() {
	// Start the server with the config and other parameters
	golog.Print("Starting Server...")
	s.Application.Run(iris.Addr(s.Port), iris.WithConfiguration(s.Configuration))
}

// SetKubeSession binds the current K8s session to the server
func (s *Server) SetKubeSession(kubeSession *kube.K8sSession) error {
	if kubeSession == nil {
		return errors.New("please provide a k8s session")
	}
	s.K8sSession = kubeSession
	return nil
}

func getDefaultIrisConfiguration() iris.Configuration {
	return iris.Configuration{
		DisableStartupLog:                 false,
		DisableInterruptHandler:           false,
		DisablePathCorrection:             false,
		EnablePathEscape:                  false,
		FireMethodNotAllowed:              false,
		DisableBodyConsumptionOnUnmarshal: false,
		DisableAutoFireStatusCode:         false,
		TimeFormat:                        "Mon, 02 Jan 2006 15:04:05 GMT",
		Charset:                           "UTF-8",
	}
}

func registerRoutes(s *Server) {
	s.Application.Get("/", indexHandler)
	s.Application.Get("/{namespace:string}/pods", getPods)
	s.Application.Get("/{namespace: string}/services", getServices)
	s.Application.Post("/stress_test/{iterations: int min(1)}", stressTestHandler)
	s.Application.Post("/create_pod/", createPodHandler)
	s.Application.Get("/info", serverInfoHandler)
}
