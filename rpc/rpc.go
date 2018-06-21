package rpc

import (
	"context"
	"time"
	"zues/config"
	pb "zues/proto"
	"zues/server"
	"zues/util"

	yaml "github.com/ghodss/yaml"
)

// RPCServer is a simple gRPC server struct
type RPCServer struct{}

// GetInfo is a rpc call to the control server to fetch the status for the server
func (s *RPCServer) GetInfo(ctx context.Context, in *pb.Empty) (*pb.InfoResponse, error) {
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

// DeployJob is a rpc call to the control to start a deployment job
func (s *RPCServer) DeployJob(ctx context.Context, jobReq *pb.JobRequest) (*pb.JobResponse, error) {
	// Get the yaml data for the job and parse
	var jobConfig config.Config
	decodedJobConfig := util.DecodeBase64(jobReq.JobDescInYaml)
	err := yaml.Unmarshal(decodedJobConfig, &jobConfig)
	if err != nil {
		return nil, err
	}
	// TODO : Create the Job and return necessary data back

	return &pb.JobResponse{
		JobID:     util.RandomString(16),
		Status:    "Created",
		CreatedAt: time.Now().Unix(),
	}, nil
}
