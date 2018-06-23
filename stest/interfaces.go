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
		startTime := time.Now().Unix()
		InMemoryTests[newStressTest.ID][i] = statisticalTelemetry{
			TestID:          st.ID,
			Name:            st.Name,
			Completed:       0,
			Remaining:       newStressTest.Spec.NumRequests,
			Total:           newStressTest.Spec.NumRequests,
			AvgResponseTime: 0,
			Success:         0,
			Status:          TestStatusCreated,
			UpdatedAt:       startTime,
			CreatedAt:       startTime,
		}
	}
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
	golog.Infof("Start stress test %s at %d", s.ID, startTime)

	// Calculate number of chunks
	nChunks := int(int(s.Spec.NumRequests) / MaxRoutineChunk)
	for i := 0; i < nChunks; i++ {
		golog.Infof("TEST: %s CHUNK: %d", s.ID, i+1)
		// Creating a channel to wait for 25 requests to finish
		chunkCompleteCh := make(chan struct{})

		// errorChan will only be updated if a invalid response code is received
		errorChan := make(chan tempStatisticalTelemetry)
		// successChan will only be updated if a successful response code is received
		successChan := make(chan tempStatisticalTelemetry)

		statUpdateDoneChan := make(chan struct{})

		// This go routine is only triggered after MaxRoutuneChunk times
		s.localTelemetry.wg.Add(1)
		go s.processWorkerResponse(statUpdateDoneChan, chunkCompleteCh)

		s.localTelemetry.wg.Add(1)
		go s.updateStatisticalTelemetry(successChan, errorChan, statUpdateDoneChan)

		// starting all the workers for the required number of times
		// successChan and errorChan are used to communicate between
		// the methods processWorkderResponse and startWorkers to
		// update the statisticalTelemetry data for each test
		s.startWorkers(chunkCompleteCh, successChan, errorChan)

		// Waiting on the computeTelemetryData go finish
		s.localTelemetry.wg.Wait()
		time.Sleep(time.Duration(s.Spec.RestDuration) * time.Millisecond)
	}
	endTime := time.Now().UnixNano()
	fmt.Printf("Elapsed time for test %s is %d ms\n", s.ID, ((endTime - startTime) / 1000000))
	golog.Println(fmt.Sprintf("In-Memory tests: %+v", InMemoryTests))
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

func (s *StressTest) processWorkerResponse(statUpdateDoneChan chan<- struct{}, chunkCompleteChan <-chan struct{}) {
	// This function only runs after the buffer on the channel is full to
	// compute statistics of the data

	//Waiting on a signal to process

	select {
	case <-chunkCompleteChan:
		// Do routine work after a chunk of work is over
		statUpdateDoneChan <- struct{}{}
		s.localTelemetry.wg.Done()
		break
	}
}

func (s *StressTest) startWorkers(chunkCompleteChan chan<- struct{}, successChan, errorChan chan<- tempStatisticalTelemetry) {
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
			// Update Test status
			inMemTest := &InMemoryTests[s.ID][st.ID-1]
			if inMemTest == nil {
				golog.Errorf("Error while fetching test %s from cache.", s.ID)
			} else if inMemTest.Status == TestStatusCreated {
				InMemoryTests[s.ID][st.ID-1].Status = TestStatusRunning
			}
			s.localTelemetry.routineWg.Add(1)
			go s.runTest(st, successChan, errorChan)
			s.localTelemetry.scheduler.Put(item)
		}
	}
	s.localTelemetry.routineWg.Wait()
	// Signal after the responses have arrived is done
	chunkCompleteChan <- struct{}{}
}

// Pass the request in as a parameter and update the queue in the startRoutines
func (s *StressTest) runTest(incomingReq *test, successChan, errorChan chan<- tempStatisticalTelemetry) {
	// Create a HTTP request
	req := buildRequest(incomingReq.Type, s.Spec.ServerPort,
		s.Spec.Selector["name"], incomingReq.Endpoint,
		incomingReq.Body, incomingReq.Headers)
	reqStartTime := time.Now().Unix()
	statusCode, _, err := util.GetHTTPResponse(req)
	elapsedTime := time.Now().Unix() - reqStartTime
	tempStTelemetry := tempStatisticalTelemetry{elapsedTime: elapsedTime, testID: incomingReq.ID}
	if err != nil {
		golog.Errorf("Error while getting response [Status Code: %d]: %s", statusCode, err.Error())
	}
	// Check to see if the response codes are correct
	if !isValidResponse(incomingReq, int16(statusCode)) {
		golog.Errorf("ID: %s Error in status code expected %v got %d", s.ID, incomingReq.ValidResponseCodes, statusCode)
		// Updating the errorChan to process error in request
		errorChan <- tempStTelemetry
	} else {
		// golog.Info(fmt.Sprintf("%s", res))
		successChan <- tempStTelemetry
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

func (s *StressTest) updateStatisticalTelemetry(successChan, errorChan <-chan tempStatisticalTelemetry, statUpdateDoneChan <-chan struct{}) {

	for {
		select {
		case exTest := <-successChan:
			//golog.Infof("Success received for test %d", exTest.testID)
			s.localTelemetry.updateMutex.Lock()
			// Lock no ensure no race condition
			sTelemetry := &InMemoryTests[s.ID][exTest.testID-1]
			// Update the parameters
			sTelemetry.Success++
			sTelemetry.Completed++
			sTelemetry.Remaining--
			// Update the average response time
			sTelemetry.AvgResponseTime += float64(exTest.elapsedTime)
			sTelemetry.AvgResponseTime += sTelemetry.AvgResponseTime / float64(sTelemetry.Total)
			// Update the timestamp to the latest timestamp
			sTelemetry.UpdatedAt = time.Now().Unix()

			// Update the status to completed if all the requests are done
			if sTelemetry.Completed == sTelemetry.Total {
				sTelemetry.Status = TestStatusCompleted
				// Calculate duration
				sTelemetry.Elapsed = time.Now().Unix() - sTelemetry.CreatedAt
			}
			s.localTelemetry.updateMutex.Unlock()
		case testID := <-errorChan:
			golog.Println(fmt.Sprintf("Error received for test %d", testID))
		case <-statUpdateDoneChan:
			// Signal the main wg that you work is done
			s.localTelemetry.wg.Done()
			break
		}
	}

}
