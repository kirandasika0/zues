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

// TestStatus represents the on going status of a test
type TestStatus string

const (
	// HTTP is the default HTTP 1.1 server
	HTTP ServerType = "http"

	// HTTPGetRequest is the HTTP verb for a GET request
	HTTPGetRequest HTTPRequestType = "GET"
	// HTTPPostRequest  is the HTTP verb for a POST request
	HTTPPostRequest HTTPRequestType = "POST"
	// HTTPDeleteRequest is the HTTP verb for a DELETE request
	HTTPDeleteRequest HTTPRequestType = "DELETE"
	// HTTPPutRequest is the HTTP verb for a PUT request
	HTTPPutRequest HTTPRequestType = "PUT"

	// TestStatusCreated signifies the current status of a Test job running in the background
	TestStatusCreated TestStatus = "Created"
	// TestStatusRunning signifies the current status of a Test job running in the background
	TestStatusRunning TestStatus = "Running"
	// TestStatusCompleted signifies the current status of a Test job running in the background
	TestStatusCompleted TestStatus = "Completed"

	// MaxResponseBuffer is a constant that is used a threshold before dumping all the response data
	MaxResponseBuffer uint32 = 1000000
	// MaxRoutineChunk is a constant that shows the number of concurrent workers per chunk
	MaxRoutineChunk int = 25
)

var (
	// InMemoryTests is a map of all tests in memory. Tests are normally removed
	// a certain time if they are not used or accessed
	InMemoryTests = map[string][]statisticalTelemetry{}
	// DispatchTestDataCh is a channel to signal the stressTestStreamDispatcher func
	DispatchTestDataCh = make(chan string)
)

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
	updateMutex    sync.RWMutex
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

// Statistical Telemetry is the main data that will
// queried from the RPC or REST calls that come to this server
// all other data will be available based on availability
type statisticalTelemetry struct {
	TestID          int16      `json:"test_id"`
	Name            string     `json:"name"`
	Completed       uint16     `json:"completed"`
	Remaining       uint16     `json:"remaining"`
	Total           uint16     `json:"total"`
	Success         int        `json:"success"`
	AvgResponseTime float64    `json:"avg_response_time"`
	Status          TestStatus `json:"status"`
	UpdatedAt       int64      `json:"updated_at"`
	CreatedAt       int64      `json:"create_at"`
	Elapsed         int64      `json:"elapsed"`
}

type tempStatisticalTelemetry struct {
	elapsedTime int64
	testID      int16
}
