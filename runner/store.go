package runner

import (
	"fmt"
	"strings"
	"time"

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

	value = strings.TrimSpace(value)
	// DynamoDB measures length via UTF-8 bytes, so does go
	if len(value) > 10000 {
		return value[:10000]
	}

	return value
}

func (d DynamoStore) writeResult(success bool, key string, result string) error {
	doc := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"Key": {
				S: aws.String(key),
			},
			"CompletedTime": {
				S: aws.String(time.Now().UTC().Format(time.RFC3339Nano)),
			},
			"Success": {
				BOOL: aws.Bool(success),
			},
		},
		TableName: aws.String(d.table),
	}

	result = d.normalizeValue(result)
	if result != "" {
		doc.Item["Result"] = &dynamodb.AttributeValue{
			S: aws.String(result),
		}
	}

	_, err := d.client.PutItem(doc)
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
		// ignore cases with empty results
		if _, ok := v["Result"]; ok {
			outputs = append(outputs, *v["Result"].S)
		}
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
