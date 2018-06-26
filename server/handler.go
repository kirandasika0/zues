package server

import (
	"fmt"
	"io/ioutil"
	"zues/config"
	pubsub "zues/dispatch"
	"zues/stest"
	"zues/util"

	"github.com/kataras/golog"
	"github.com/kataras/iris"
)

func indexHandler(ctx iris.Context) {
	configStr, err := ioutil.ReadFile("zues-config.yaml")
	if err != nil {
		golog.Error(err)
	}
	zuesBaseConfig, err := config.GetConfigFromYAML(configStr)
	if err != nil {
		golog.Error(err)
	}

	util.BuildResponse(ctx, zuesBaseConfig)
}

func getPods(ctx iris.Context) {}

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
