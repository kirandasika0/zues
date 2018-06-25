package main

import (
	"context"
	"io/ioutil"
	"os"
	"time"
	pb "zues/proto"
	"zues/util"

	"github.com/kataras/golog"

	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("137.135.124.197:80", grpc.WithInsecure())
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
		os.Exit(1)
	}

	for i := 0; i < 20; i++ {
		ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		jDetails, err := c.JobDetails(ctx, &pb.JobRequest{
			JobID: res.JobID,
		})
		if err != nil {
			golog.Error(err)
			break
		}
		golog.Infof("%+v", jDetails)
		time.Sleep(2 * time.Second)
	}

	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err = c.DeleteJob(ctx, &pb.JobRequest{
		JobID: res.JobID,
	})
	if err != nil {
		golog.Error(err)
		os.Exit(1)
	}
	golog.Infof("JobID: %s deleted", res.JobID)
}
