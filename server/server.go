package server

import (
	"net"
	"net/http"
	pubsub "zues/dispatch"
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
		CheckOrigin:     allowOrigins,
	}
	// All the allowed client to connect to this server as a websocket
	allowedWsOrigins = []string{
		"http://localhost:8284",
		"http://137.135.124.197",
		"file://",
	}
)

func allowOrigins(r *http.Request) bool {
	// TODO: Make this parametersized so that we can only allow certain client
	// estabilish the websocket connection.
	requestOrigin := r.Header["Origin"][0]
	golog.Warnf("Attempt to establish websocket connection from %s", requestOrigin)
	for _, o := range allowedWsOrigins {
		if requestOrigin == o {
			golog.Infof("Success in establising websocket connection by client %s", requestOrigin)
			return true
		}
	}
	golog.Errorf("Error establishing websocket connection by client %s", requestOrigin)
	return false
}

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

	// Handle streaming testData back to the connected websockets
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
	// server level routes
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
	s.Application.Get("/test/{job_id: string}/logs/stream", jobLogStreamHandler)
	s.Application.Get("/test/status/stream/{job_id: string}", stressTestStatusStreamHandler)
}

// stressTestStreamDispatcher is a helper func that listens on the DispatchTestDataCh
// (which is triggered only when an entire stress test is completed) and broadcasts
// the statisticalTelemetryData to the websocket connections
func stressTestStreamDispatcher() {
	for {
		select {
		case jobID := <-stest.DispatchTestDataCh:
			c, err := pubsub.GetChannel(jobID)
			if err != nil {
				golog.Errorf(err.Error())
				continue
			}
			// Broadcasting the message to the listerns now
			tests, _ := stest.InMemoryTests[jobID]
			err = c.Broadcast(tests)
			if err != nil {
				golog.Errorf(err.Error())
			}
		}
	}
}
