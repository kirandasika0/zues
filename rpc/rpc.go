package rpc

import (
	"context"
	"fmt"
	"time"
	"zues/config"
	"zues/kube"
	pb "zues/proto"
	"zues/server"
	"zues/util"

	"github.com/kataras/golog"

	yaml "github.com/ghodss/yaml"
)

// GRPCServer is a simple gRPC server struct
type GRPCServer struct{}

// GetInfo is a rpc call to the control server to fetch the status for the server
func (s *GRPCServer) GetInfo(ctx context.Context, in *pb.Empty) (*pb.InfoResponse, error) {
	return &pb.InfoResponse{
		Port:     server.ZuesServer.Port,
		ServerID: server.ZuesServer.ServerID,
	}, nil
}

// DeployJob is a rpc call to the control to start a deployment job
func (s *GRPCServer) DeployJob(ctx context.Context, jobReq *pb.JobRequest) (*pb.JobResponse, error) {
	// Get the yaml data for the job and parse
	var jobConfig config.Config
	decodedJobConfig := util.DecodeBase64(jobReq.JobDescInYaml)
	err := yaml.Unmarshal(decodedJobConfig, &jobConfig)
	if err != nil {
		return nil, err
	}
	// TODO : Create the Job and return necessary data back
	JobID := "job" + "-" + util.RandomString(16)
	golog.Infof("JobID: %s CREATED", JobID)

	container := kube.CreateContainer("docker", "hello-world")
	_, err = kube.Session.CreatePod(jobConfig.Spec.Name, jobConfig.Spec.Namespace, nil, container)
	if err != nil {
		return nil, err
	}
	// Add to the in Memory jobs
	config.CurrentJobs[JobID] = jobConfig
	return &pb.JobResponse{
		JobID:     JobID,
		Status:    "Created",
		CreatedAt: time.Now().Unix(),
	}, nil
}

// JobDetails is a rpc call
func (s *GRPCServer) JobDetails(ctx context.Context, req *pb.JobRequest) (*pb.JobDetailResponse, error) {
	job, ok := config.CurrentJobs[req.JobID]
	if !ok {
		return nil, fmt.Errorf("Could not locate JobID: %s", req.JobID)
	}
	res := &pb.JobDetailResponse{
		JobID:          req.JobID,
		JobStatus:      "Running...",
		MaxBuildErrors: job.Spec.MaxBuildErrors,
		MaxRetries:     job.Spec.MaxRetries,
	}
	return res, nil
}
