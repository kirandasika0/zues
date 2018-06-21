package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"
	pb "zues/proto"
	"zues/util"

	"github.com/kataras/golog"

	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:8284", grpc.WithInsecure())
	if err != nil {
		golog.Error(err)
	}
	defer conn.Close()

	c := pb.NewZuesControlClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	content, _ := ioutil.ReadFile("../zues-config.yaml")
	encodedBytes := util.EncodeBase64(content)
	jobRequest := &pb.JobRequest{JobDescInYaml: encodedBytes, Timestamp: time.Now().Unix()}
	res, err := c.DeployJob(ctx, jobRequest)
	if err != nil {
		golog.Error(err)
	}
	golog.Info(fmt.Sprintf("%+v", res))
}
