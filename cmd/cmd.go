package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Clever/batchcli/runner"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func main() {
	functionCmd := flag.String("cmd", "", "The command to run")
	localRun := flag.Bool("local", false, "Local mode - auto-assigns random AWS Batch config")
	printVersion := flag.Bool("version", false, "Print the version and exit")
	resultsLocation := flag.String("results-location", "workflows-dev-infra", "name of the table for getting and setting job results")
	//parseArgs := flag.Bool("parseargs", true, "If false send the job payload directly to the cmd as its first argument without parsing it")

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

	store, err := initalizeStore(*localRun, *resultsLocation)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if *resultsLocation == "" {
		fmt.Println("-results-location can not be an empty string")
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

func initalizeStore(localRun bool, resultsLocation string) (runner.ResultsStore, error) {
	// TODO: use fake dynamo for localRun
	config := aws.NewConfig().WithRegion("us-east-1")
	config.Credentials = credentials.NewCredentials(&credentials.ChainProvider{
		VerboseErrors: false,
		Providers: []credentials.Provider{
			&credentials.SharedCredentialsProvider{Filename: "", Profile: ""},
			defaults.RemoteCredProvider(*config, defaults.Handlers()),
		},
	})
	sess, err := session.NewSession(config)

	if err != nil {
		return runner.DynamoStore{}, nil
	}
	// override the default credential chain to disable envaccesskeys
	client := dynamodb.New(sess)

	return runner.NewDynamoStore(client, resultsLocation)
}
