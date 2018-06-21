package main

import (
	"net"
	"os"
	"zues/kube"
	pb "zues/proto"
	"zues/rpc"
	"zues/server"

	"github.com/kataras/golog"
	"google.golang.org/grpc"

	"github.com/soheilhy/cmux"
	"golang.org/x/sync/errgroup"
)

func main() {
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

	// Start the K8s session
	kube.Session = kube.New()

	m := cmux.New(listener)
	grpcListener := m.Match(cmux.HTTP2HeaderField("content-type", "application/grpc"))
	httpListener := m.Match(cmux.HTTP1Fast())

	g := new(errgroup.Group)
	g.Go(func() error { return servegRPCServer(grpcListener) })
	g.Go(func() error { return serveHTTPServer(httpListener) })
	g.Go(func() error { return m.Serve() })

	golog.Println("Running servers")
	g.Wait()
}

func serveHTTPServer(l net.Listener) error {
	server.ZuesServer = server.New(nil, "")
	server.ZuesServer.Start(l)
	return nil
}

func servegRPCServer(l net.Listener) error {
	golog.Println("Starting gRPC server...")
	s := grpc.NewServer()
	pb.RegisterZuesControlServer(s, &rpc.GRPCServer{})
	return s.Serve(l)
}
