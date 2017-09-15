package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/Clever/batchcli/runner"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sfn"
)

func main() {
	activityName := flag.String("name", "", "The activity name to register with AWS Step Functions")
	activityCmd := flag.String("cmd", "", "The command to run to process activity tasks")
	sfnRegion := flag.String("region", "", "The AWS region to send Step Function API calls")
	printVersion := flag.Bool("version", false, "Print the version and exit")

	flag.Parse()

	if *printVersion {
		//fmt.Println(Version)
		os.Exit(0)
	}

	if *activityName == "" {
		fmt.Println("activityname is required")
		os.Exit(1)
	}
	if *activityCmd == "" {
		fmt.Println("cmd is required")
		os.Exit(1)
	}
	if *sfnRegion == "" {
		fmt.Println("region is required")
		os.Exit(1)
	}

	ctx := context.Background()

	// register the activity with AWS (it might already exist, which is ok)
	sfnapi := sfn.New(session.New(), aws.NewConfig().WithRegion(*sfnRegion))
	createOutput, err := sfnapi.CreateActivityWithContext(ctx, &sfn.CreateActivityInput{
		Name: activityName,
	})
	if err != nil {
		fmt.Printf("error creating activity: %s\n", err)
		os.Exit(1)
	} else {
		fmt.Printf("running as activity: %s\n", *createOutput.ActivityArn)
	}

	// run getactivitytask and get some work
	// getactivitytask itself claims to initiate a polling loop, but wrap it in a polling loop of our own
	// since it seems to return every minute or so with a nil error and empty output
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ctx.Done():
			fmt.Println("terminating GetActivityTask polling loop")
			break
		case <-ticker.C:
			getATOutput, err := sfnapi.GetActivityTaskWithContext(ctx, &sfn.GetActivityTaskInput{
				ActivityArn: createOutput.ActivityArn,
				// WorkerName: "TODO: something useful about this process: host? pid?",
			})
			if err != nil {
				fmt.Printf("error getting activity task: %s\n", err)
				os.Exit(1)
			}
			if getATOutput.TaskToken == nil {
				fmt.Println("nil task token starting GetActivityTask again")
				continue
			}
			fmt.Println("got activity task", getATOutput)
			activityTaskInput := getATOutput.Input
			activityTaskToken := getATOutput.TaskToken

			// start a heartbeat TODO
			//  taskContext, cancelFn := context.WithCancel()
			// go func() {
			// 	heartbeat := time.NewTicker(10 * time.Second)
			// 	for {
			// 		select {
			// 		case <-taskContext.Done():
			// 			fmt.Println("ending heartbeat loop")
			// 			return
			// 		case <-hearbeat.C:
			// 			_, err := sfnapi.
			// 		}
			// 	}
			// }()

			taskRunner := runner.NewTaskRunner(*activityCmd, flag.Args(), sfnapi, *activityTaskInput, *activityTaskToken)
			if err := taskRunner.Process(ctx); err != nil {
				fmt.Printf("error running process: %s\n", err)
				continue
			}
		}
	}
}
