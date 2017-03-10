package runner

import (
	"errors"
	"os"
	"strings"
)

type BatchJob struct {
	JobId              string   // Required: AWS Batch JobId for this Job
	Queue              string   // Required: AWS Batch Queue Name this Job was posted in
	ComputeEnvironment string   // Required: AWS Batch Cluster Name running this container
	DependencyIds      []string // Optional: JobIds that are dependencies for this Job run
	Input              string   // Optional: Input data recieved from workflow-manager
}

// NewBatchJobFromEnv returns a new BatchJob by reading
// values from the environment. Assumes this code is running
// via AWS Batch
func NewBatchJobFromEnv() (BatchJob, error) {
	// Read environment variables injected by AWS Batch
	jobId := os.Getenv("AWS_BATCH_JOB_ID")
	queue := os.Getenv("AWS_BATCH_JQ_NAME")
	computeEnv := os.Getenv("AWS_BATCH_CE_NAME")

	// Should be all present if this Job is running via AWS Batch
	if jobId == "" || queue == "" || computeEnv == "" {
		return BatchJob{},
			errors.New("AWS Batch environment variables not found")
	}

	// Optional env var with job-ids of dependencies
	// Required if job results from dependencies need to be fetched
	depsStr := os.Getenv("_BATCH_DEPENDENCIES")
	dependencyIds := []string{}
	if depsStr != "" {
		dependencyIds = strings.Split(depsStr, ",")
	}

	// Workflow input is passed as a string through this env var
	input := os.Getenv("_BATCH_START")

	return BatchJob{
		JobId:              jobId,
		Queue:              queue,
		ComputeEnvironment: computeEnv,
		DependencyIds:      dependencyIds,
		Input:              input,
	}, nil
}

// NewMockBatchJob returns a new BatchJob with
// mock values for tests and local runs
func NewMockBatchJob(deps []string) BatchJob {
	return BatchJob{
		JobId:              "local",
		Queue:              "fake-queue",
		ComputeEnvironment: "local",
		DependencyIds:      deps,
	}
}
