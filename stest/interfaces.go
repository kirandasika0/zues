package stest

import (
	"fmt"
	"sync"
	"zues/util"

	"github.com/kataras/golog"
)

var wg = sync.WaitGroup{}
var completeCount = 0
var buffer []string

// SetupStressTestEnvironment sets up the stress test environment. Sets up go routines
// and other locks to ensure no race conditions
func SetupStressTestEnvironment(server string, endpoint string, iterations int) []string {

	if len(server) == 0 {
		golog.Error("need a server to ping")
	}

	if len(buffer) > 0 {
		buffer = []string{}
		completeCount = 0
	}

	// bodyDataCh will act as a data channel between the runSingleStressTest go routine
	// and the computeStatistics go routine
	bodyDataCh := make(chan []byte)
	// running computeStatistics as a go routine go avoid deadlock
	go computeStatistics(bodyDataCh)

	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go runSingleStressTest(server, endpoint, bodyDataCh)
	}

	wg.Wait()
	golog.Print(fmt.Sprintf("Completed %d stress test on url: %s", completeCount+1, server+endpoint))
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
