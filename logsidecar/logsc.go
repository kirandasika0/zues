package main

import (
	"context"
	"log"
	"net"
	"os"
	proto "zues/logsidecar/logsidecar"

	"google.golang.org/grpc"
)

type server struct{}

func (s *server) GetStatus(ctx context.Context, v *proto.Void) (*proto.SidecarStatus, error) {
	log.Print("Getting status...")
	return &proto.SidecarStatus{}, nil
}

func (s *server) ConfigureSidecar(ctx context.Context, config *proto.SidecarBasicConfig) (*proto.SidecarStatus, error) {
	return nil, nil
}

func main() {
	networkString := "localhost:49449"
	if os.Getenv("DOCKER_ENV") != "" {
		networkString = "0.0.0.0:" + os.Getenv("SIDECAR_PORT")
	}
	l, err := net.Listen("tcp", networkString)
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer()
	proto.RegisterSidecarServer(s, &server{})
	s.Serve(l)
}
