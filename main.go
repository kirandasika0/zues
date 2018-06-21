package main

import (
	"context"
	"net"
	"zues/kube"
	"zues/server"

	"github.com/kataras/golog"
	"google.golang.org/grpc"

	pb "zues/proto"

	"github.com/soheilhy/cmux"
	"golang.org/x/sync/errgroup"
)

type rpcServer struct{}

func (s *rpcServer) GetInfo(ctx context.Context, in *pb.Empty) (*pb.InfoResponse, error) {
	return &pb.InfoResponse{
		Port:     server.ZuesServer.Port,
		ServerID: server.ZuesServer.ServerID,
		K8SSession: &pb.K8SSession{
			ServerAddress: server.ZuesServer.K8sSession.ServerAddress,
			ServerPort:    uint32(server.ZuesServer.K8sSession.ServerPort),
			ServerBaseUrl: server.ZuesServer.K8sSession.ServerBaseURL,
			AccessToken:   server.ZuesServer.K8sSession.AccessToken,
			ApiCalls:      server.ZuesServer.K8sSession.APICalls,
		},
	}, nil
}

func main() {
	listener, err := net.Listen("tcp", "localhost:8284")
	if err != nil {
		panic(err)
	}
	m := cmux.New(listener)
	grpcListener := m.Match(cmux.HTTP2HeaderField("content-type", "application/grpc"))
	httpListener := m.Match(cmux.HTTP1Fast())

	g := new(errgroup.Group)
	g.Go(func() error { return serveGRPCServer(grpcListener) })
	g.Go(func() error { return serveHTTPServer(httpListener) })
	g.Go(func() error { return m.Serve() })

	golog.Println("Running servers")
	g.Wait()
}

func serveHTTPServer(l net.Listener) error {
	server.ZuesServer = server.New(nil, "")
	kube.K8sGlobalSession = kube.New()
	server.ZuesServer.SetKubeSession(kube.K8sGlobalSession)
	server.ZuesServer.Start(l)
	return nil
}

func serveGRPCServer(l net.Listener) error {
	golog.Println("Starting gRPC server...")
	grpcServer := grpc.NewServer()
	pb.RegisterZuesControlServer(grpcServer, &rpcServer{})
	return grpcServer.Serve(l)
}
