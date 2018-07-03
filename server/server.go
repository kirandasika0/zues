package server

import (
	"net"
	"net/http"
	"os"
	pubsub "zues/dispatch"
	"zues/stest"
	"zues/util"

	"github.com/gorilla/websocket"
	"github.com/kataras/golog"
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	recover2 "github.com/kataras/iris/middleware/recover"
	"github.com/rs/xid"
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
	LogFile       string             `json:"log_file"`
}

// New creates a new instance of the zues server
func New(serverConfig *iris.Configuration, serverPort, version string) *Server {
	var s Server
	if serverConfig == nil {
		s.Configuration = getDefaultIrisConfiguration()
	}

	// Start up a new Iris Application
	s.Application = iris.New()
	s.Application.Use(recover2.New())
	s.Application.UseGlobal(before)

	if len(serverPort) < 1 {
		// Set default port to localhost 8284
		s.Port = ":8284"
	} else {
		s.Port = serverPort
	}

	s.ServerID = "zues-master-" + util.RandomString(8)
	s.Health = "ok"

	// Log config
	logConfig := logger.Config{
		Status:            true,
		IP:                true,
		Method:            true,
		Path:              true,
		MessageHeaderKeys: []string{"X-Trace-Id"},
	}
	customLogger := logger.New(logConfig)
	s.Application.Use(customLogger)
	s.Application.Logger().SetLevel("debug")
	if os.Getenv("IN_CLUSTER") != "" {
		file, fileName, err := logFile(s.ServerID, version)
		s.LogFile = fileName
		if err != nil {
			golog.Errorf("Could not create a log file")
		}
		s.Application.Logger().AddOutput(file)
	}
	// Register all the routes to the server
	registerRoutes(&s)

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
	s.Application.Get("/healthz", livelinessProbeHandler)
	s.Application.Post("/logs/trigger/upload", logsUploadHandler)
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

func before(ctx iris.Context) {
	// Check for trace header
	_, ok := ctx.Request().Header["X-Trace-Id"]
	if !ok {
		ctx.Request().Header.Set("X-Trace-Id", xid.New().String())
	}
	ctx.Next()
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

// logFile opens a new log file and returns it to the main control
func logFile(serverID, version string) (*os.File, string, error) {
	var file *os.File
	fileName := "/var/log/" + serverID + "-" + version + ".log"
	if os.Getenv("IN_CLUSTER") != "" {
		// TODO: create the file in the log volume mount
		file, _ = os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	} else {
		file, _ = os.OpenFile(serverID+"-"+version+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	}
	return file, fileName, nil
}
