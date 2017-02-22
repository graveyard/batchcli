package runner

import (
	"bytes"
	"io"
	"os"
	"os/exec"

	"gopkg.in/Clever/kayvee-go.v6/logger"
)

var log = logger.New("batchcli")

type TaskRunner struct {
	job    BatchJob
	store  ResultsStore
	cmd    []string
	inputs []string
}

// Process runs the underlying cmd with the appropriate
// environment and command line params
func (t TaskRunner) Process() error {
	log.InfoD("exec-command", map[string]interface{}{
		"inputs": t.inputs,
		"cmd":    t.cmd,
		"job-id": t.job.JobId,
	})

	args := append(t.cmd[1:], t.inputs...)
	cmd := exec.Command(t.cmd[0:1], args...)

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

func NewTaskRunner(cmd []string, job BatchJob, store ResultsStore) (TaskRunner, error) {
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
