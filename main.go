package main

import (
	"zues/kube"
	"zues/server"
)

func main() {
	server.ZuesServer = server.New(nil, "")
	kube.K8sGlobalSession = kube.New()
	server.ZuesServer.SetKubeSession(kube.K8sGlobalSession)
	server.ZuesServer.Start()
}
