package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/aws/aws-sdk-go/service/sfn/sfniface"

	"gopkg.in/Clever/kayvee-go.v6/logger"
)

var log = logger.New("sfncli")

type TaskRunner struct {
	sfnapi    sfniface.SFNAPI
	taskToken string
	cmd       string
	inputs    []string
}

// Process runs the underlying cmd with the appropriate
// environment and command line params
func (t TaskRunner) Process(ctx context.Context) error {
	if t.sfnapi == nil {
		return nil // if New failed :-/
	}
	log.InfoD("exec-command", map[string]interface{}{
		"inputs": t.inputs,
		"cmd":    t.cmd,
	})

	cmd := exec.CommandContext(ctx, t.cmd, t.inputs...)
	cmd.Env = os.Environ()

	// Write the stdout and stderr of the process to both this process' stdout and stderr
	// and also write to a byte buffer so that we can save it in the ResultsStore
	var stderrbuf bytes.Buffer
	var stdoutbuf bytes.Buffer
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrbuf)
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutbuf)

	if err := cmd.Run(); err != nil {
		if _, e := t.sfnapi.SendTaskFailureWithContext(ctx, &sfn.SendTaskFailureInput{
			Cause:     aws.String(stderrbuf.String()), // TODO: limits on length?
			TaskToken: &t.taskToken,
		}); e != nil {
			return fmt.Errorf("error sending task failure: %s", e)
		}
		return err
	}

	// output must be JSON if it isn't make it so
	output := stdoutbuf.String()
	var test interface{}
	if err := json.Unmarshal([]byte(output), &test); err != nil {
		// output isn't JSON, make it json
		newOutputBs, _ := json.Marshal(map[string]interface{}{
			"raw": output,
		})
		output = string(newOutputBs)
	}

	_, err := t.sfnapi.SendTaskSuccessWithContext(ctx, &sfn.SendTaskSuccessInput{
		Output:    aws.String(output), // TODO: limits on length?
		TaskToken: &t.taskToken,
	})
	return err
}

func NewTaskRunner(cmd string, args []string, sfnapi sfniface.SFNAPI, taskInput string, taskToken string) TaskRunner {
	var params []string
	if err := json.Unmarshal([]byte(taskInput), &params); err != nil {
		sfnapi.SendTaskFailure(&sfn.SendTaskFailureInput{
			Cause:     aws.String(fmt.Sprintf("Task input must be array of strings: %s", err.Error())),
			TaskToken: &taskToken,
		})
		return TaskRunner{}
	}

	// append the input on the cmd passed through the CLI
	// example:
	// 		sfncli -cmd echo how now
	// 		input = ["brown", "cow"]
	//      exec(echo, ["how", "now", "brown", "cow"])
	inputs := append(args, params...)

	return TaskRunner{
		sfnapi:    sfnapi,
		taskToken: taskToken,
		cmd:       cmd,
		inputs:    inputs,
	}
}
