package runner

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type ResultsStore interface {
	Success(key, result string) error
	Failure(key, result string) error
	GetResults(keys []string) ([]string, error)
}

type DynamoStore struct {
	client *dynamodb.DynamoDB
	table  string
}

func (d DynamoStore) writeResult(success bool, key string, result string) error {
	_, err := d.client.PutItem(&dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"Key": {
				S: aws.String(key),
			},
			"Result": {
				S: aws.String(result),
			},
			"Success": {
				BOOL: aws.Bool(true),
			},
		},
		TableName: aws.String(d.table),
	})

	if err != nil {
		return err
	}
	return nil
}

func (d DynamoStore) Success(key string, result string) error {
	return d.writeResult(true, key, result)
}

func (d DynamoStore) Failure(key string, result string) error {
	return d.writeResult(false, key, result)
}

func (d DynamoStore) GetResults(keys []string) ([]string, error) {
	return []string{}, nil
}

func NewDynamoStore(table string) (ResultsStore, error) {
	sess, err := session.NewSession()
	if err != nil {
		return DynamoStore{}, err
	}

	client := dynamodb.New(sess)

	return DynamoStore{
		client,
		table,
	}, nil
}
