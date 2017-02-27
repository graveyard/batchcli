package runner

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type ResultsStore interface {
	Success(key, result string) error
	Failure(key, result string) error
	GetResults(keys []string) ([]string, error)
}

type DynamoStore struct {
	client dynamodbiface.DynamoDBAPI
	table  string
}

// normalizeValue returns a truncated value if the length
// of the input value is greater than 128 bytes
func (d DynamoStore) normalizeValue(value string) string {
	// TODO: maybe we also should enforce that values are JSON-only
	// for now just enforce a length

	// DynamoDB measures length via UTF-8 bytes, so does go
	if len(value) > 128 {
		return value[:128]
	}

	return value
}

func (d DynamoStore) writeResult(success bool, key string, result string) error {
	_, err := d.client.PutItem(&dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"Key": {
				S: aws.String(key),
			},
			"Result": {
				S: aws.String(d.normalizeValue(result)),
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
	if len(keys) == 0 {
		return []string{}, nil
	}

	fetchKeys := []map[string]*dynamodb.AttributeValue{}
	for _, k := range keys {
		fetchKeys = append(fetchKeys, map[string]*dynamodb.AttributeValue{
			"Key": {S: aws.String(k)},
		})
	}

	results, err := d.client.BatchGetItem(&dynamodb.BatchGetItemInput{
		RequestItems: map[string]*dynamodb.KeysAndAttributes{
			d.table: {
				Keys: fetchKeys,
			},
		},
	})
	if err != nil {
		return []string{}, err
	}

	// TODO: we should keep a sense of job-id and the success.
	// For now just pull out the responses
	if _, ok := results.Responses[d.table]; !ok {
		return []string{}, fmt.Errorf("No response for keys %s in table: %s", keys, d.table)
	}

	var outputs = []string{}
	for _, v := range results.Responses[d.table] {
		outputs = append(outputs, *v["Result"].S)
	}

	return outputs, nil
}

func NewDynamoStore(client dynamodbiface.DynamoDBAPI, table string) (ResultsStore, error) {
	if table == "" {
		return DynamoStore{}, fmt.Errorf("table name must be a non-empty string")
	}

	return DynamoStore{
		client,
		table,
	}, nil
}
