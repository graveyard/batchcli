package runner

import (
	"bytes"
	"io"
	"os"
	"os/exec"
)

type TaskRunner struct {
	job    BatchJob
	store  ResultsStore
	cmd    string
	inputs []string
}

// Process runs the underlying cmd with the appropriate
// environment and command line params
func (t TaskRunner) Process() error {
	cmd := exec.Command(t.cmd, t.inputs...)
	cmd.Env = os.Environ()

	// Write the stdout and stderr of the process to both this process' stdout and stderr
	// and also write it to a byte buffer so that we can return it with the Gearman job
	// data as necessary.
	var stderrbuf bytes.Buffer
	var stdoutbuf bytes.Buffer
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrbuf)
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutbuf)

	if err := cmd.Run(); err != nil {
		t.store.Failure(t.job.JobId, stderrbuf.String())
		return err
	}

	t.store.Success(t.job.JobId, stdoutbuf.String())
	return nil
}

func NewTaskRunner(cmd string, job BatchJob, store ResultsStore) (TaskRunner, error) {
	inputs, err := store.GetResults(job.DependencyIds)
	if err != nil {
		return TaskRunner{}, err
	}

	return TaskRunner{
		job,
		store,
		cmd,
		inputs,
	}, nil
}
