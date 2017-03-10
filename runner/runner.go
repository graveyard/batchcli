package runner

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/Clever/kayvee-go.v6/logger"
)

var log = logger.New("batchcli")

type TaskRunner struct {
	job    BatchJob
	store  ResultsStore
	cmd    string
	inputs []string
}

// Process runs the underlying cmd with the appropriate
// environment and command line params
func (t TaskRunner) Process() error {
	log.InfoD("exec-command", map[string]interface{}{
		"inputs":       t.inputs,
		"cmd":          t.cmd,
		"job-id":       t.job.JobId,
		"dependencies": strings.Join(t.job.DependencyIds, ","),
	})

	cmd := exec.Command(t.cmd, t.inputs...)
	cmd.Env = os.Environ()

	// Write the stdout and stderr of the process to both this process' stdout and stderr
	// and also write to a byte buffer so that we can save it in the ResultsStore
	var stderrbuf bytes.Buffer
	var stdoutbuf bytes.Buffer
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrbuf)
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutbuf)

	if err := cmd.Run(); err != nil {
		if e := t.store.Failure(t.job.JobId, stderrbuf.String()); e != nil {
			return fmt.Errorf("Failed to write failure: %s. reason: %s", err, e)
		}
		return err
	}

	return t.store.Success(t.job.JobId, stdoutbuf.String())
}

func NewTaskRunner(cmd string, args []string, job BatchJob, store ResultsStore) (TaskRunner, error) {
	results, err := store.GetResults(job.DependencyIds)
	if err != nil {
		return TaskRunner{}, err
	}
	// postfix the results of previous jobs on the cmd passed through
	// the CLI
	// example:
	// 		batchcli -cmd echo hello there
	// 		results = [{"json":"true"}, {}]
	//      exec(echo, ["hello", "there", '{"json":"true"}', '{}'])
	inputs := append(args, results...)

	return TaskRunner{
		job,
		store,
		cmd,
		inputs,
	}, nil
}
