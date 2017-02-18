package runner

import (
	"errors"
	"os"
	"strings"
)

type BatchJob struct {
	JobId              string
	Queue              string
	ComputeEnvironment string
	DependencyIds      []string
}

// NewBatchJobFromEnv returns a new BatchJob by reading
// values from the envrionment
func NewBatchJobFromEnv() (BatchJob, error) {
	jobId := os.Getenv("AWS_BATCH_JOB_ID")
	queue := os.Getenv("AWS_BATCH_JQ_NAME")
	computeEnv := os.Getenv("AWS_BATCH_CE_NAME")

	if jobId == "" || queue == "" || computeEnv == "" {
		return BatchJob{},
			errors.New("AWS Batch environment variables not found")
	}

	depsStr := os.Getenv("_BATCH_DEPEDENCIES")
	dependencyIds := []string{}
	if depsStr != "" {
		dependencyIds = strings.Split(depsStr, ",")
	}

	return BatchJob{
		JobId:              jobId,
		Queue:              queue,
		ComputeEnvironment: computeEnv,
		DependencyIds:      dependencyIds,
	}, nil
}
