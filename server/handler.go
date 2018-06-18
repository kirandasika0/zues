package server

import (
	"fmt"
	"io/ioutil"
	"zues/config"
	"zues/kube"
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

func getPods(ctx iris.Context) {
	namespace := ctx.Params().Get("namespace")
	body, err := util.GetHTTPBody(kube.APIServer, fmt.Sprintf("/api/v1/namespaces/%s/pods", namespace))
	if err != nil {
		golog.Error(err)
	}

	pods, err := ZuesServer.K8sSession.GetPodsFromAPIServer(body)
	if err != nil {
		golog.Error(err)
	}
	if len(pods) == 0 {
		util.BuildErrorResponse(ctx, err.Error())
	} else {
		util.BuildResponse(ctx, pods)
	}
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
	}

	// Initiate the environment
	err = stressTest.InitStressTestEnvironment()
	if err != nil {
		util.BuildErrorResponse(ctx, err.Error())
	}
	// Execute the environment
	stressTest.ExecuteEnvironment()

	util.BuildResponse(ctx, stressTest)
}

func getServices(ctx iris.Context) {
	namespace := ctx.Params().Get("namespace")
	body, err := util.GetHTTPBody(kube.APIServer, fmt.Sprintf("/api/v1/namespaces/%s/services", namespace))
	if err != nil {
		golog.Error(err)
	}
	services, err := ZuesServer.K8sSession.GetServicesFromAPIServer(body)
	if err != nil {
		golog.Error(err)
	}
	util.BuildResponse(ctx, services)
}

func createPodHandler(ctx iris.Context) {
	requestData, err := util.ExtractHTTPBody(ctx.Request())
	if err != nil {
		golog.Error(err)
		util.BuildErrorResponse(ctx, err.Error())
		return
	}

	pod, err := ZuesServer.K8sSession.CreatePodWithNamespace(requestData, "sprintt-qa")

	if err != nil {
		util.BuildErrorResponse(ctx, err.Error())
		return
	}
	util.BuildResponse(ctx, pod)
}

func serverInfoHandler(ctx iris.Context) {
	util.BuildResponse(ctx, ZuesServer)
}

func deletePodHandler(ctx iris.Context) {
	podName := ctx.Params().Get("podName")
	namespace := ctx.Params().Get("namespace")
	uid := ctx.Params().Get("uid")
	if len(podName) < 1 || len(namespace) < 1 {
		util.BuildErrorResponse(ctx, "need a pod name to delete.")
		return
	}

	pod, err := ZuesServer.K8sSession.DeletePodWithNamespace(podName, namespace, uid)
	if err != nil {
		util.BuildErrorResponse(ctx, err.Error())
		return
	}

	util.BuildResponse(ctx, pod)
}
