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
	container := kube.CreateContainer(jobConfig.Spec.Name, jobConfig.Spec.Image, 8001)
	podLabels := map[string]string{"app": "candidate-service", "run": "candidate-service"}
	pod, err := kube.Session.CreatePod(jobConfig.Spec.Name,
		jobConfig.Spec.Namespace,
		podLabels,
		container)
	if err != nil {
		return nil, err
	}
	golog.Infof("JobID: %s CREATED", JobID)
	// Add to the in Memory jobs
	config.CurrentJobs[JobID] = jobConfig

	// Add to the in Memory Pods for getting real time status
	config.JobPodsMap[JobID] = pod

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

	jobPod, ok := config.JobPodsMap[req.JobID]
	if !ok {
		return nil, fmt.Errorf("Could not locate the Pod for Job ID: %s", req.JobID)
	}

	var restarts int32
	for _, c := range jobPod.Status.ContainerStatuses {
		restarts += c.RestartCount
	}

	jobResponse := &pb.JobDetailResponse{
		JobID:          req.JobID,
		JobStatus:      string(jobPod.Status.Phase),
		MaxBuildErrors: job.Spec.MaxBuildErrors,
		MaxRetries:     job.Spec.MaxRetries,
		ErrorsOccured:  restarts,
	}
	golog.Infof("JobID: %s Pod: %s Status: %s", req.JobID, jobPod.ObjectMeta.Name, jobPod.Status.Phase)
	return jobResponse, nil
}

// DeleteJob is a rpc method to delete a current job in the system
func (s *GRPCServer) DeleteJob(ctx context.Context, req *pb.JobRequest) (*pb.Empty, error) {
	if req.JobID == "" {
		return nil, fmt.Errorf("error: please provide a JobID to search")
	}
	_, ok := config.CurrentJobs[req.JobID]
	if !ok {
		return nil, fmt.Errorf("Could not locate JobID: %s", req.JobID)
	}

	jobPod, ok := config.JobPodsMap[req.JobID]
	if !ok {
		return nil, fmt.Errorf("No pod found for JobID: %s", req.JobID)
	}

	kube.Session.DeletePod(jobPod.ObjectMeta.Name, jobPod.ObjectMeta.Namespace)

	delete(config.CurrentJobs, req.JobID)
	delete(config.JobPodsMap, req.JobID)

	return &pb.Empty{}, nil
}
