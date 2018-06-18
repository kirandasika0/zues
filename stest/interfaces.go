package stest

import (
	"errors"
	"fmt"
	"sync"
	"time"
	"zues/util"

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
	return &newStressTest, nil
}

// InitStressTestEnvironment does some required checks and sets up
// a stress testing environment
func (s *StressTest) InitStressTestEnvironment() error {

	s.localTelemetry.wg = sync.WaitGroup{}

	// SERVICE DISCOVERY
	// First estabilish a TCP connection and check to see if service is running

	/*
		* TODO change from localhost to acutal service name in future
		* Use the selector name from the pod as the service name default
		unless specified
	*/
	if !util.HasTCPConnection("localhost", fmt.Sprintf(":%d", s.Spec.ServerPort)) {
		return errors.New("Unable to contact host server. Tried to establish a TCP connection")
	}

	// Queue up all the requests in order
	// Scheduler used throughout the testing phase
	scheduler := queue.New(int64(len(s.Spec.Tests)))

	// Pre-process all data into the right format

	for _, test := range s.Spec.Tests {

		// All Authorization values are provided in base64 encoding
		for k, v := range test.Authorization {
			test.Authorization[k] = string(util.DecodeBase64(v))
		}

		if test.Type == HTTPPostRequest {
			// Decode everything from base64
			test.Body = string(util.DecodeBase64(test.Body))
		}
	}

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
	fmt.Printf("Started test: %s at: %d\n", s.ID, startTime)

	// Creating a buffered channel to wait for 25 requests to finish
	chunkCompleteCh := make(chan struct{})
	s.localTelemetry.wg.Add(1)
	// This go routine is only triggered after MaxRoutuneChunk times
	go s.computeTelemetryData(chunkCompleteCh)

	// Create routines to send HTTP requests
	// for i := 0; i < 26; i++ {
	// 	s.localTelemetry.wg.Add(1)
	// 	go s.runTest(chunkCompleteCh)
	// }
	s.startRoutines(chunkCompleteCh)
	endTime := time.Now().UnixNano()
	fmt.Printf("Elapsed time for test %s is %d ns", s.ID, (endTime - startTime))

	// Waiting on the computeTelemetryData go finish
	s.localTelemetry.wg.Wait()

	time.Sleep(time.Duration(s.Spec.RestDuration) * time.Millisecond)
	close(chunkCompleteCh)
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

func (s *StressTest) computeTelemetryData(chunkChan <-chan struct{}) {
	// This function only runs after the buffer on the channel is full to
	// compute statistics of the data
	for range chunkChan {
		fmt.Println("computing...")
	}

	s.localTelemetry.wg.Done()
}

func (s *StressTest) startRoutines(chunkChan chan<- struct{}) {
	for i := 0; i < MaxRoutineChunk; i++ {
		s.localTelemetry.wg.Add(1)
		go s.runTest()
	}
	s.localTelemetry.wg.Wait()
	// Signal after the time is done
	chunkChan <- struct{}{}
}

func (s *StressTest) runTest() {
	fmt.Println("running...")
	s.localTelemetry.wg.Done()
}
