package server

import (
	"net"
	"zues/stest"
	"zues/util"

	"github.com/gorilla/websocket"
	"github.com/kataras/golog"
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	recover2 "github.com/kataras/iris/middleware/recover"
)

var (
	// ZuesServer Global ZuesServer instance
	ZuesServer *Server

	// Upgrader handles upgrading the HTTP request to a websocket connection
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	// JobListeners maps the a jobID to a slice of listeners
	JobListeners = map[string][]*websocket.Conn{}
)

// Server holds all the necessary information for the zues HTTP API to function
type Server struct {
	Application   *iris.Application  `json:"-"`
	Port          string             `json:"port"`
	ServerID      string             `json:"serverId"`
	Configuration iris.Configuration `json:"-"`
	Health        string             `json:"health"`
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

	s.ServerID = "zues-master-" + util.RandomString(8)
	s.Health = "ok"

	go stressTestStreamDispatcher()

	return &s
}

// Start starts the zues server
func (s *Server) Start(l net.Listener) {
	// Start the server with the config and other parameters
	golog.Print("Starting Server...")
	s.Application.Run(iris.Listener(l))
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
	// Application level routes
	s.Application.Get("/", indexHandler)
	s.Application.Get("/info", serverInfoHandler)
	// kube package routes
	s.Application.Get("/jobs", listJobsHandler)
	s.Application.Get("/{namespace:string}/pods", getPods)
	s.Application.Get("/{namespace: string}/services", getServices)
	s.Application.Post("/{namespace: string}/pod/", createPodHandler)
	s.Application.Delete("/pod/{namespace: string}/{podName :string}/{uid: string}", deletePodHandler)

	// Stress test package routes
	s.Application.Get("/tests", listTestHandler)
	s.Application.Post("/test", stressTestHandler)
	s.Application.Get("/test/status/{test_id: string}", stressTestStatusHandler)

	// Stream handlers
	s.Application.Get("/test/status/stream/{job_id: string}", stressTestStatusStreamHandler)
}

func stressTestStreamDispatcher() {
	for {
		select {
		case jobID := <-stest.DispatchTestDataCh:
			conns, ok := JobListeners[jobID]
			if !ok {
				break
			}
			for _, conn := range conns {
				websocket.WriteJSON(conn, stest.InMemoryTests[jobID])
			}
		}
	}
}
