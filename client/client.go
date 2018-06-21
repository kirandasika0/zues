package main

import (
	"context"
	"io/ioutil"
	"sync"
	"time"
	pb "zues/proto"
	"zues/util"

	"github.com/kataras/golog"

	"google.golang.org/grpc"
)

var wg = sync.WaitGroup{}

func main() {
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go getResponse()
	}
	wg.Wait()
}

func getResponse() {
	conn, err := grpc.Dial("localhost:8284", grpc.WithInsecure())
	if err != nil {
		golog.Error(err)
	}
	defer conn.Close()

	c := pb.NewZuesControlClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	content, _ := ioutil.ReadFile("../zues-config.yaml")
	encodedBytes := util.EncodeBase64(content)
	jobRequest := &pb.JobRequest{JobDescInYaml: encodedBytes, Timestamp: time.Now().Unix()}
	res, err := c.DeployJob(ctx, jobRequest)
	if err != nil {
		golog.Error(err)
	}
	jDetails, err := c.JobDetails(ctx, &pb.JobRequest{
		JobID: res.JobID,
	})
	golog.Infof("%+v", jDetails)
	wg.Done()
}
