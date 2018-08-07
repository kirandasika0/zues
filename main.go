package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
	"zues/kube"
	pb "zues/proto/server"
	"zues/rpc"
	"zues/server"

	"github.com/kataras/golog"
	"google.golang.org/grpc"

	"github.com/soheilhy/cmux"
)

var (
	// Version is used to point at the type of build that is being run; default is development
	Version = "dev"

	// BuildTime is used to point at the time this code is compiled
	BuildTime = fmt.Sprintf("%d", time.Now().Unix())

	// GitSHA will be linked during compilation
	GitSHA = ""

	listenAddr *string
)

func init() {
	listenAddr = flag.String("addr", "localhost:8284", "addr is the unix host and port")

	flag.Parse()
}

func main() {
	golog.Infof("zues =====>  version:%s pid: %d build: %s", Version, os.Getpid(), BuildTime)

	// Docker needs to use the 0.0.0.0 format to forward all requests
	// to the server in the container
	listener, err := net.Listen("tcp", *listenAddr)
	if err != nil {
		panic(err)
	}

	startWatch := make(chan struct{})

	// Start the K8s session
	kube.Session, err = kube.New()
	if err != nil {
		golog.Error(err.Error())
	} else {
		// Start watching after servers startup
		go kube.Session.WatchPodEvents(startWatch)
		// Send start signal after 2 seconds
		go func() {
			select {
			case <-time.After(2 * time.Second):
				startWatch <- struct{}{}
			}
		}()
	}

	m := cmux.New(listener)
	grpcListener := m.Match(cmux.HTTP2HeaderField("content-type", "application/grpc"))
	httpListener := m.Match(cmux.HTTP1Fast())

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)

	go serveHTTPServer(httpListener)
	go servegRPCServer(grpcListener)
	go m.Serve()
	<-c

	golog.Debug("Shutdown initiated")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err = server.ZuesServer.Application.Shutdown(ctx); err != nil {
		golog.Error("error while shutting down gracefully...")
		golog.Debug("falling back to hard terminate")
		os.Exit(1)
	}
	os.Exit(0)
}

func serveHTTPServer(l net.Listener) error {
	server.ZuesServer = server.New(nil, "", Version)
	server.ZuesServer.Start(l)
	return nil
}

func servegRPCServer(l net.Listener) error {
	s := grpc.NewServer()
	pb.RegisterZuesControlServer(s, &rpc.GRPCServer{})
	return s.Serve(l)
}
