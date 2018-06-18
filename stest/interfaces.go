package stest

import (
	"errors"
	"fmt"
	"sync"
	"time"
	"zues/util"

	"github.com/Workiva/go-datastructures/queue"
	"github.com/kataras/golog"
)

var wg = sync.WaitGroup{}
var completeCount = 0
var buffer []string

// SetupStressTestEnvironment sets up the stress test environment. Sets up go routines
// and other locks to ensure no race conditions
func SetupStressTestEnvironment(server string, endpoint string, iterations int) []string {

	if len(buffer) > 0 {
		buffer = []string{}
		completeCount = 0
	}

	// bodyDataCh will act as a data channel between the runSingleStressTest go routine
	// and the computeStatistics go routine
	bodyDataCh := make(chan []byte)
	// running computeStatistics as a go routine go avoid deadlock
	go computeStatistics(bodyDataCh)

	// Calculate start time
	startTime := int(time.Now().UTC().UnixNano())

	// If lees than 100 requests to perform just run everything at once.
	if iterations <= 100 {
		for i := 0; i < iterations; i++ {
			wg.Add(1)
			go runSingleStressTest(server, endpoint, bodyDataCh)
		}
	} else {
		// TODO  Release the goroutines in batches
		for i := 0; i < iterations; i++ {
			wg.Add(1)
			go runSingleStressTest(server, endpoint, bodyDataCh)
		}
	}

	wg.Wait()
	endTime := int(time.Now().UTC().UnixNano())
	elapsedTime := endTime - startTime
	golog.Print(fmt.Sprintf("Completed %d stress test on url: %s took %d ms", completeCount+1, server+endpoint, (elapsedTime / 1000000)))
	close(bodyDataCh)
	return buffer
}

func runSingleStressTest(server string, endpoint string, bodyDataCh chan<- []byte) {
	defer wg.Done()
	body, err := util.GetHTTPBody(server, endpoint)
	if err != nil {
		golog.Error(err)
	}
	bodyDataCh <- body
}

func computeStatistics(bodyDataCh <-chan []byte) {
	for bodyBuffer := range bodyDataCh {
		buffer = append(buffer, string(bodyBuffer))
		completeCount++
	}
}

// InitateStressTestEnvironment does some required checks and sets up
// a stress testing environment
func (s *StressTest) InitateStressTestEnvironment() error {
	// SERVICE DISCOVERY
	// first estabilish a TCP connection and check
	// if the server is running

	// TODO change from localhost to acutal service name in future
	if !util.HasTCPConnection("localhost", fmt.Sprintf(":%d", s.Spec.ServerPort)) {
		return errors.New("Unable to contact host server. Tried to establish a TCP connection")
	}

	// Queue up all the requests in order (Priority Queue)
	// Start the timer
	// Chunk the go routines into chunks of 25
	// Using a channel to signal after every 25 requests chan struct{}
	// Nested go routines
	// Level 1 is for number of tests
	// Level 2 is for number of requests per test (ie running 25 requests per cycle)
	// Dump buffer is it exceeds the MaxResponseBuffer and reset executionTrace
	// if needed to save memory
	scheduler := queue.New(int64(len(s.Spec.Tests)))
	// Queueing up all the jobs needed
	for _, test := range s.Spec.Tests {
		scheduler.Put(test)
	}
	firstItem, err := scheduler.Peek()
	if err != nil {
		panic("problem")
	}
	fmt.Printf("%v\n", firstItem)
	return nil
}
