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

	util.BuildResponse(ctx, zuesBaseConfig)
}

func getPods(ctx iris.Context) {
	namespace := ctx.Params().Get("namespace")
	body, err := util.GetHTTPBody(kube.APIServer, fmt.Sprintf("/api/v1/namespaces/%s/pods", namespace))
	if err != nil {
		golog.Error(err)
	}

	pods, err := kube.GetPodsFromJSON(body)
	if err != nil {
		golog.Error(err)
	}
	if len(pods) == 0 {
		util.BuildResponse(ctx, map[string]string{"error": err.Error()})
	} else {
		util.BuildResponse(ctx, pods)
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
	util.BuildResponse(ctx, largeBuffer)
}

func getServices(ctx iris.Context) {
	namespace := ctx.Params().Get("namespace")
	body, err := util.GetHTTPBody(kube.APIServer, fmt.Sprintf("/api/v1/namespaces/%s/services", namespace))
	if err != nil {
		golog.Error(err)
	}
	services, err := kube.GetServicesFromAPIServer(body)
	if err != nil {
		golog.Error(err)
	}
	util.BuildResponse(ctx, services)
}

func createPodHandler(ctx iris.Context) {
	var requestData map[string]interface{}
	ctx.ReadJSON(&requestData)
	// Acccess the K8s API server to create a pod with the given spec
	req, err := util.CreateHTTPRequest("POST", "http://localhost:8001/api/v1/namespaces/default/pods",
		map[string]string{"Content-Type": "application/json"}, requestData)
	if err != nil {
		util.BuildResponse(ctx, map[string]string{"error": err.Error()})
		return
	}

	_, resp, err := util.GetHTTPResponse(req)
	if err != nil {
		util.BuildResponse(ctx, map[string]string{"error": err.Error()})
	}

	util.BuildResponse(ctx, resp)
}
