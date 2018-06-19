package stest

import (
	"net/http"
	"net/http/httptrace"
	"sync"

	"github.com/Workiva/go-datastructures/queue"
)

// ServerType is a abstract type for a server
// Defining all the types of servers.
// Currently only support for HTTP1.1 servers
type ServerType string

// HTTPRequestType signifies the various requests present
type HTTPRequestType string

type TestStatus string

const (
	// HTTP is the default HTTP 1.1 server
	HTTP ServerType = "http"

	HTTPGetRequest    HTTPRequestType = "GET"
	HTTPPostRequest   HTTPRequestType = "POST"
	HTTPDeleteRequest HTTPRequestType = "DELETE"
	HTTPPutRequest    HTTPRequestType = "PUT"

	TestStatusCreated  TestStatus = "Created"
	TestStatusRunning  TestStatus = "Running"
	TestStatusComplted TestStatus = "Completed"

	MaxResponseBuffer uint32 = 1000000
	MaxRoutineChunk   int    = 25
)

// InMemoryTests is a map of all tests in memory. Tests are normally removed
// a certain time if they are not used or accessed
var InMemoryTests = map[string]interface{}{}

// StressTest struct defines the parameters need for the stress test
type StressTest struct {
	ID             string         `json:"id,omitempty"`
	APIVersion     string         `json:"apiVersion"`
	Kind           string         `json:"kind"`
	Spec           stressTestSpec `json:"spec"`
	Notify         bool           `json:"notify"`
	localTelemetry executionTelemetry
}

type stressTestSpec struct {
	Selector       map[string]string `json:"selector"`
	NumRequests    uint16            `json:"numRequests"`
	NumConcurrent  int16             `json:"numConcurrent"`
	RestDuration   int               `json:"restDuration"`
	ServerType     ServerType        `json:"serverType"`
	ServerPort     int16             `json:"serverPort"`
	Tests          []test            `json:"tests"`
	ExecutionOrder []int16           `json:"executionOrder"`
}

type test struct {
	ID       int16           `json:"id"`
	Name     string          `json:"name"`
	Type     HTTPRequestType `json:"type"`
	Endpoint string          `json:"endpoint"`
	//the Body must always be encoding in base64
	// Linux command echo '<your-text>' | base64
	Body               string  `json:"body,omitempty"`
	ValidResponseCodes []int16 `json:"validResponseCodes"`
	// Authorization value should also be provided in base64 encoding
	Authorization map[string]string `json:"auth"`
	Headers       map[string]string `json:"headers,omitempty"`
}

type executionTelemetry struct {
	wg             sync.WaitGroup
	routineWg      sync.WaitGroup
	scheduler      *queue.Queue
	executionQueue *queue.Queue
	ElapsedTime    int        `json:"elapsed_time"`
	Status         TestStatus `json:"status"`
	ResponseBuffer []string   `json:"response_buffer,omitempty"`
}

type stressTestRequest struct {
	request *http.Request
	trace   *httptrace.ClientTrace
}
