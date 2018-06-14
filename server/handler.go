package server

import (
	"zues/util"
	"github.com/kataras/iris"
	"github.com/kataras/golog"
	"zues/config"
	"zues/kube"
	"fmt"
	"strconv"
	"zues/stest"
	"io/ioutil"
	"encoding/json"
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
