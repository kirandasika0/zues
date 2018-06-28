package main

import (
	"fmt"
	"net"
	"os"
	"time"
	"zues/kube"
	pb "zues/proto"
	"zues/rpc"
	"zues/server"

	"github.com/kataras/golog"
	"google.golang.org/grpc"

	"github.com/soheilhy/cmux"
	"golang.org/x/sync/errgroup"
)

// Replaced to approprate value with -ldflags
var version = "development"

func main() {
	fmt.Printf("running version %s\n", version)
	// Docker needs to use the 0.0.0.0 format to forward all requests
	// to the server in the container
	var networkString = "localhost:8284"
	if os.Getenv("DOCKER_ENV") == "true" {
		networkString = "0.0.0.0:8284"
	}
	listener, err := net.Listen("tcp", networkString)
	if err != nil {
		panic(err)
	}

	startWatch := make(chan struct{})

	// Start the K8s session
	kube.Session = kube.New()
	// Start watching after servers startup
	go kube.Session.WatchPodEvents(startWatch)

	m := cmux.New(listener)
	grpcListener := m.Match(cmux.HTTP2HeaderField("content-type", "application/grpc"))
	httpListener := m.Match(cmux.HTTP1Fast())

	g := new(errgroup.Group)
	g.Go(func() error { return servegRPCServer(grpcListener) })
	g.Go(func() error { return serveHTTPServer(httpListener) })
	g.Go(func() error { return m.Serve() })

	// Start watching on the Pod events
	go func() {
		select {
		case <-time.After(2 * time.Second):
			startWatch <- struct{}{}
		}
	}()
	g.Wait()
}

func serveHTTPServer(l net.Listener) error {
	golog.Info("Starting HTTP server...")
	server.ZuesServer = server.New(nil, "")
	server.ZuesServer.Start(l)
	return nil
}

func servegRPCServer(l net.Listener) error {
	golog.Info("Starting gRPC server...")
	s := grpc.NewServer()
	pb.RegisterZuesControlServer(s, &rpc.GRPCServer{})
	return s.Serve(l)
}
