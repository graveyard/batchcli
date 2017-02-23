package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Clever/batchcli/runner"
)

func main() {
	functionCmd := flag.String("cmd", "", "The command to run")
	localRun := flag.Bool("local", false, "Local mode - auto-assigns random AWS Batch config")
	//parseArgs := flag.Bool("parseargs", true, "If false send the job payload directly to the cmd as its first argument without parsing it")

	printVersion := flag.Bool("version", false, "Print the version and exit")

	flag.Parse()

	if *printVersion {
		fmt.Println(Version)
		os.Exit(0)
	}

	if *functionCmd == "" {
		fmt.Println("Required command to execute")
		os.Exit(1)
	}

	var job runner.BatchJob
	if *localRun {
		// TODO: allow CLI to set fake job-ids (or inputs)
		job = runner.NewMockBatchJob([]string{})
	} else {
		j, err := runner.NewBatchJobFromEnv()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		job = j
	}

	// TODO: use fake dynamo for localRun
	store, err := runner.NewDynamoStore("test-batch-workflows")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	taskRunner, err := runner.NewTaskRunner(*functionCmd, flag.Args(), job, store)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := taskRunner.Process(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}
