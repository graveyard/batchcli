# batchcli

Utility to manage inputs and outputes of batch worker executions. Uses DynamoDB as backend

Owned by eng-infra

## AWS Policy

A policy of the following form be added to the _ECS Task Role_:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "Stmt1490140880000",
            "Effect": "Allow",
            "Action": [
                "dynamodb:*"
            ],
            "Resource": [
                "arn:aws:dynamodb:us-east-1:<account-id>:table/workflow-results*"
            ]
        }
    ]
}
```

The name of the Dyanmo table can be changed using the `-results-location` flag. It defaults to `workflow-results-dev`.

## Usage

```
batchcli -cmd <command> cmd-args
```

## Shepherds

- "mohit.gupta@clever.com"
