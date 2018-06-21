package main

import (
	"context"
	"fmt"
	"time"
	pb "zues/proto"

	"github.com/kataras/golog"

	"google.golang.org/grpc"
)

func main() {
	for {
		conn, err := grpc.Dial("localhost:8284", grpc.WithInsecure())
		if err != nil {
			golog.Error(err)
		}
		defer conn.Close()

		c := pb.NewZuesControlClient(conn)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		res, err := c.GetInfo(ctx, &pb.Empty{})
		if err != nil {
			golog.Error(err)
		}
		golog.Info(fmt.Sprintf("%+v", res.K8SSession.ApiCalls))

		time.Sleep(2 * time.Second)
	}
}
