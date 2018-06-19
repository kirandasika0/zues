package stest

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
	"zues/util"

	"github.com/kataras/golog"

	"github.com/Workiva/go-datastructures/queue"
	yaml "github.com/ghodss/yaml"
)

// New return an instance of the StressTest struct
func New(config []byte) (*StressTest, error) {
	var newStressTest StressTest
	err := yaml.Unmarshal(config, &newStressTest)
	if err != nil {
		return nil, err
	}
	// Create the new ID and return
	newStressTest.ID = util.RandomString(16)

	// Pre-process all data into the right format
	for _, test := range newStressTest.Spec.Tests {

		// All Authorization values are provided in base64 encoding
		for k, v := range test.Authorization {
			test.Authorization[k] = string(util.DecodeBase64(v))
		}
	}

	// Add it to the InMemoryTests
	if _, ok := InMemoryTests[newStressTest.ID]; !ok {
		InMemoryTests[newStressTest.ID] = make([]statisticalTelemetry, len(newStressTest.Spec.Tests))
	}
	for i, st := range newStressTest.Spec.Tests {
		InMemoryTests[newStressTest.ID][i] = statisticalTelemetry{
			TestID:    st.ID,
			Completed: 0,
			Remaining: newStressTest.Spec.NumRequests,
			Success:   0,
			Status:    TestStatusCreated,
			Timestamp: time.Now().Unix(),
		}
	}
	golog.Println(InMemoryTests)
	return &newStressTest, nil
}

// InitStressTestEnvironment does some required checks and sets up
// a stress testing environment
func (s *StressTest) InitStressTestEnvironment() error {

	s.localTelemetry.wg = sync.WaitGroup{}
	s.localTelemetry.routineWg = sync.WaitGroup{}
	s.localTelemetry.updateMutex = sync.RWMutex{}
	s.localTelemetry.executionQueue = queue.New(int64(MaxRoutineChunk))

	// SERVICE DISCOVERY
	// First estabilish a TCP connection and check to see if service is running

	/*
		* TODO change from localhost to acutal service name in future
		* Use the selector name from the pod as the service name default
		unless specified
	*/
	if s.Spec.ServerPort == 0 {
		// Defaulting to 80 as HTTP incoming port
		s.Spec.ServerPort = 80
	}
	if !util.HasTCPConnection(s.Spec.Selector["name"], fmt.Sprintf(":%d", s.Spec.ServerPort)) {
		return errors.New("Unable to contact host server. Tried to establish a TCP connection")
	}

	// Queue up all the requests in order
	// Scheduler used throughout the testing phase
	scheduler := queue.New(int64(len(s.Spec.Tests)))

	for _, testID := range s.Spec.ExecutionOrder {
		testToQueue, err := s.findTest(testID)
		if err != nil {
			return err
		}

		scheduler.Put(testToQueue)
	}

	// Attach scheduler to the StressTest
	s.localTelemetry.scheduler = scheduler

	return nil
}

// ExecuteEnvironment sets up the stress test environment. Sets up go routines
// and other locks to ensure no race conditions
func (s *StressTest) ExecuteEnvironment() {
	// Start the timer
	// Chunk the go routines into chunks of 25
	// Using a channel to signal after every 25 requests chan struct{}
	// Nested go routines
	// Level 1 is for number of tests
	// Level 2 is for number of requests per test (ie running 25 requests per cycle)
	// Dump buffer is it exceeds the MaxResponseBuffer and reset executionTrace
	// if needed to save memory
	startTime := time.Now().UnixNano()
	fmt.Printf("Start stress test %s at %d\n", s.ID, startTime)

	// Calculate number of chunks
	nChunks := int(int(s.Spec.NumRequests) / MaxRoutineChunk)
	for i := 0; i < nChunks; i++ {
		fmt.Printf("TEST: %s CHUNK: %d\n", s.ID, i+1)
		// Creating a buffered channel to wait for 25 requests to finish
		chunkCompleteCh := make(chan struct{})
		s.localTelemetry.wg.Add(1)
		// This go routine is only triggered after MaxRoutuneChunk times
		go s.processChunkResponse(chunkCompleteCh)

		s.startRoutines(chunkCompleteCh)

		// Waiting on the computeTelemetryData go finish
		s.localTelemetry.wg.Wait()
		time.Sleep(time.Duration(s.Spec.RestDuration) * time.Millisecond)
	}
	endTime := time.Now().UnixNano()
	fmt.Printf("Elapsed time for test %s is %d ms\n", s.ID, ((endTime - startTime) / 1000000))
}

func (s *StressTest) findTest(targetTestID int16) (*test, error) {

	if targetTestID == 0 {
		return nil, errors.New("Please provide a valid target id")
	}

	for _, t := range s.Spec.Tests {
		if targetTestID == t.ID {
			return &t, nil
		}
	}

	return nil, fmt.Errorf("Could not find the test with id: %d", targetTestID)
}

func (s *StressTest) processChunkResponse(chunkChan <-chan struct{}) {
	// This function only runs after the buffer on the channel is full to
	// compute statistics of the data

	//Waiting on a signal to process
	<-chunkChan
	qLen := s.localTelemetry.executionQueue.Len()

	items, err := s.localTelemetry.executionQueue.Get(qLen)
	if err != nil {

	}
	for _, item := range items {
		_, ok := item.(*test)
		if !ok {
			panic("type convertion failed")
		}
		s.updateStatisticalTelemetry()
	}

	s.localTelemetry.wg.Done()
}

func (s *StressTest) startRoutines(chunkChan chan<- struct{}) {
	for i := 0; i < MaxRoutineChunk; i++ {
		qLen := s.localTelemetry.scheduler.Len()
		items, err := s.localTelemetry.scheduler.Get(qLen)
		if err != nil {
			golog.Error(err.Error())
		}
		for _, item := range items {
			st, ok := item.(*test)
			if !ok {
				golog.Errorf("Error in type conversion.")
			}
			s.localTelemetry.routineWg.Add(1)
			go s.runTest(st)
			s.localTelemetry.scheduler.Put(item)
		}
	}
	s.localTelemetry.routineWg.Wait()
	// Signal after the responses have arrived is done
	chunkChan <- struct{}{}
}

// Pass the request in as a parameter and update the queue in the startRoutines
func (s *StressTest) runTest(incomingReq *test) {
	// Create a HTTP request
	req := buildRequest(incomingReq.Type, s.Spec.ServerPort,
		s.Spec.Selector["name"], incomingReq.Endpoint,
		incomingReq.Body, incomingReq.Headers)
	statusCode, res, err := util.GetHTTPResponse(req)
	if err != nil {
		golog.Errorf("Error while getting response [Status Code: %d]: %s", statusCode, err.Error())
	}
	// Check to see if the response codes are correct
	if !isValidResponse(incomingReq, int16(statusCode)) {
		golog.Errorf("ID: %s Error in status code expected %v got %d", s.ID, incomingReq.ValidResponseCodes, statusCode)
	} else {
		golog.Info(fmt.Sprintf("%s", res))
	}
	// Update queue
	s.localTelemetry.executionQueue.Put(incomingReq)
	s.localTelemetry.routineWg.Done()
}

func buildRequest(method HTTPRequestType, port int16, server, endpoint string, body string, headers map[string]string) *http.Request {
	// Check for port
	url := fmt.Sprintf("http://%s:%d%s", server, port, endpoint)
	if port == 0 {
		// Connecting to default port
		url = fmt.Sprintf("http://%s%s", server, endpoint)
	}
	var req *http.Request
	var err error
	if method == HTTPGetRequest {
		req, err = util.CreateHTTPRequest(string(method), url, headers, nil)
	} else if method == HTTPPostRequest {
		payloadData, err := base64.StdEncoding.DecodeString(body)
		if err != nil {
			panic(err)
		}
		req, err = util.CreateHTTPRequest(string(method), url, headers, payloadData)
	}
	if err != nil {
		golog.Errorf("Error in create a %s HTTP request for URL: %s", method, url)
	}
	return req
}

func isValidResponse(executedTest *test, responseStatusCode int16) bool {
	for _, validCode := range executedTest.ValidResponseCodes {
		if responseStatusCode == validCode {
			return true
		}
	}
	return false
}

func (s *StressTest) updateStatisticalTelemetry() {
	s.localTelemetry.updateMutex.Lock()
	s.localTelemetry.updateMutex.Unlock()
}
