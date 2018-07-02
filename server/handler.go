package server

import (
	"context"
	"fmt"
	"time"
	"zues/config"
	pubsub "zues/dispatch"
	"zues/kube"
	log_sidecar "zues/proto/logsidecar"
	"zues/stest"
	"zues/util"

	"github.com/kataras/golog"
	"github.com/kataras/iris"
	"google.golang.org/grpc"
)

func indexHandler(ctx iris.Context) { util.BuildResponse(ctx, ZuesServer) }

func getPods(ctx iris.Context) {}

func livelinessProbeHandler(ctx iris.Context) {
	hasLivenessProbeHeader := ctx.Request().Header.Get("X-Liveness-Probe-Test")
	if hasLivenessProbeHeader != "" {
		ctx.StatusCode(iris.StatusOK)
		ctx.Write([]byte("ok"))
		return
	}

	ctx.StatusCode(iris.StatusInternalServerError)
}

func stressTestHandler(ctx iris.Context) {
	// Procedure to set up stress tests
	// 1. Read the POST body for the incoming request
	// 2. Convert from YAML to JSON
	// 3. Unmarshal to the StressTest struct and use it

	body, err := util.ExtractHTTPBody(ctx.Request())
	if err != nil {
		util.BuildErrorResponse(ctx, err.Error())
		return
	}

	// Create an instance of the stress test environment
	stressTest, err := stest.New(body)
	if err != nil {
		util.BuildErrorResponse(ctx, err.Error())
		return
	}

	// Initiate the environment
	err = stressTest.InitStressTestEnvironment()
	if err != nil {
		util.BuildErrorResponse(ctx, err.Error())
		return
	}
	// Execute the environment
	go stressTest.ExecuteEnvironment()

	util.BuildResponse(ctx, stressTest)
}

func stressTestStatusHandler(ctx iris.Context) {
	testID := ctx.Params().Get("test_id")
	value, ok := stest.InMemoryTests[testID]
	if !ok {
		util.BuildErrorResponse(ctx, fmt.Sprintf("Error test with id %s not found", testID))
		return
	}

	util.BuildResponse(ctx, value)
}

func listTestHandler(ctx iris.Context) { util.BuildResponse(ctx, stest.InMemoryTests) }

func getServices(ctx iris.Context) {}

func createPodHandler(ctx iris.Context) {}

func serverInfoHandler(ctx iris.Context) { util.BuildResponse(ctx, ZuesServer) }

func deletePodHandler(ctx iris.Context) {}

func stressTestStatusStreamHandler(ctx iris.Context) {
	jobID := ctx.Params().Get("job_id")

	_, ok := stest.InMemoryTests[jobID]
	if !ok {
		golog.Errorf("job with id %s not found", jobID)
		util.BuildErrorResponse(ctx, fmt.Sprintf("job with id %s not found", jobID))
		return
	}
	// Upgrade connection if jobID is found in CurrentJobs
	wsConn, err := upgrader.Upgrade(ctx.ResponseWriter(), ctx.Request(), nil)
	if err != nil {
		golog.Error(err.Error())
		util.BuildErrorResponse(ctx, err.Error())
		return
	}

	// Checking and/or creating a channel
	var c *pubsub.Channel
	c, err = pubsub.GetChannel(jobID)
	if c != nil {
		c.AddListener(wsConn)
	} else {
		c, err = pubsub.NewChannel(jobID)
		if err != nil {
			golog.Errorf("Error: %s", err.Error())
		}
		c.AddListener(wsConn)
	}
}

func listJobsHandler(ctx iris.Context) {
	util.BuildResponse(ctx, config.CurrentJobs)
}

// TODO: This function should handle streaming all the logs to the client
func jobLogStreamHandler(ctx iris.Context) {
	jobID := ctx.Params().Get("job_id")
	// Check if the job is in memory
	// _, ok := stest.InMemoryTests[jobID]
	// if !ok {
	// 	util.BuildErrorResponse(ctx, fmt.Sprintf("Job with id %s not found", jobID))
	// 	return
	// }
	wsConn, err := upgrader.Upgrade(ctx.ResponseWriter(), ctx.Request(), nil)
	if err != nil {
		golog.Error(err.Error())
		util.BuildErrorResponse(ctx, err.Error())
		return
	}
	// Invoke log tracking here
	go kube.Session.StreamLogsToChannel("candidate-service", jobID+"-logs-stream", wsConn)
}

func logsUploadHandler(ctx iris.Context) {
	golog.Debug("Need to trigger a log upload")
	c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	networkString := "localhost:49449"
	conn, err := grpc.Dial(networkString, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	client := log_sidecar.NewSidecarClient(conn)
	res, err := client.GetStatus(c, &log_sidecar.Void{})
	if err != nil {
		panic(err)
	}
	golog.Debugf("%v", res)
}
