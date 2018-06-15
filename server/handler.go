package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
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

	util.BuildResponse(ctx, zuesBaseConfig, false)
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
		util.BuildResponse(ctx, map[string]string{"error": err.Error()}, true)
	} else {
		util.BuildResponse(ctx, pods, false)
	}
}

func stressTestHandler(ctx iris.Context) {
	iterations, err := strconv.Atoi(ctx.Params().Get("iterations"))
	if err != nil {
		golog.Error(err)
	}
	var requestBody util.ZuesRequestBody
	ctx.ReadJSON(&requestBody)
	buffer := stest.SetupStressTestEnvironment(requestBody.Data, "/", iterations)

	var largeBuffer []interface{}
	for _, b := range buffer {
		var item interface{}
		json.Unmarshal([]byte(b), &item)
		largeBuffer = append(largeBuffer, item)
	}
	util.BuildResponse(ctx, largeBuffer, false)
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
	util.BuildResponse(ctx, services, false)
}

func createPodHandler(ctx iris.Context) {
	requestData, err := util.ExtractHTTPBody(ctx.Request())
	if err != nil {
		golog.Error(err)
		util.BuildResponse(ctx, map[string]string{"error": err.Error()}, true)
		return
	}

	pod, err := ZuesServer.K8sSession.CreatePodWithNamespace(requestData, "sprintt-qa")

	if err != nil {
		util.BuildResponse(ctx, map[string]string{"error": err.Error()}, true)
		return
	}
	util.BuildResponse(ctx, pod, false)
}

func serverInfoHandler(ctx iris.Context) {
	util.BuildResponse(ctx, ZuesServer, false)
}

func deletePodHandler(ctx iris.Context) {
	podName := ctx.Params().Get("podName")
	namespace := ctx.Params().Get("namespace")
	uid := ctx.Params().Get("uid")
	if len(podName) < 1 || len(namespace) < 1 {
		util.BuildResponse(ctx, map[string]string{"error": "need a pod name to delete."}, true)
		return
	}

	pod, err := ZuesServer.K8sSession.DeletePodWithNamespace(podName, namespace, uid)
	if err != nil {
		util.BuildResponse(ctx, map[string]string{"error": err.Error()}, true)
		return
	}

	util.BuildResponse(ctx, pod, false)
}
