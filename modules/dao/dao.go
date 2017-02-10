// Copyright 2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//	http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package dao

import (
	"fmt"
	"strconv"

	cfnclient "github.com/awslabs/ecs-secrets/modules/cloudformation/client"
	ddbclient "github.com/awslabs/ecs-secrets/modules/dynamodb/client"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

//go:generate mockgen.sh github.com/awslabs/ecs-secrets/modules/dao DAO mock/dao_mock.go

// SecretRecord defines the payload used to interact with DynamoDB
type SecretRecord struct {
	Name             string
	Serial           int64
	EncryptedData    string
	EncryptedDataKey string
	Active           bool
}

// DAO defines the interface to interact with the Data Access Layer for accessing secrets
type DAO interface {
	GetLatestVersion(string) (*SecretRecord, error)
	GetSecretRecord(string, int64) (*SecretRecord, error)
	PutSecretRecord(*SecretRecord) error
	RevokeSecretRecord(string, int64) error
}

type dao struct {
	appName        string
	dynamodbClient ddbclient.Client
}

// NewDAO creates a new DAO object backed by DynamoDB
func NewDAO(appName string, dynamodbClient ddbclient.Client) DAO {
	return &dao{
		appName:        appName,
		dynamodbClient: dynamodbClient,
	}
}

// GetSecretRecord gets a secret record from DynamoDB
func (d *dao) GetSecretRecord(namespace string, serial int64) (*SecretRecord, error) {
	key := map[string]*dynamodb.AttributeValue{
		"Name": {
			S: aws.String(namespace),
		},
		"Serial": {
			N: aws.String(strconv.FormatInt(serial, 10)),
		},
	}

	itemResult, err := d.dynamodbClient.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(cfnclient.GetSecretsTableName(d.appName)),
		Key:       key,
	})

	if err != nil {
		return nil, err
	}

	if itemResult == nil || itemResult.Item == nil || len(itemResult.Item) == 0 {
		return nil, fmt.Errorf("Secret record not found in the data store")
	}

	loadedSecret := &SecretRecord{}
	err = dynamodbattribute.UnmarshalMap(itemResult.Item, loadedSecret)

	if err != nil {
		return nil, err
	}

	return loadedSecret, nil
}

// PutSecretRecord puts a secret record into DynamoDB
func (d *dao) PutSecretRecord(record *SecretRecord) error {
	item, err := dynamodbattribute.MarshalMap(*record)
	if err != nil {
		return err
	}
	_, err = d.dynamodbClient.PutItem(&dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String(cfnclient.GetSecretsTableName(d.appName)),
	})
	return err
}

// RevokeSecretRecord revokes a secret record in DynamoDB.
func (d *dao) RevokeSecretRecord(namespace string, serial int64) error {
	_, err := d.dynamodbClient.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String(cfnclient.GetSecretsTableName(d.appName)),
		Key: map[string]*dynamodb.AttributeValue{
			"Name":   &dynamodb.AttributeValue{S: aws.String(namespace)},
			"Serial": &dynamodb.AttributeValue{N: aws.String(strconv.FormatInt(serial, 10))},
		},
		AttributeUpdates: map[string]*dynamodb.AttributeValueUpdate{
			"Active": &dynamodb.AttributeValueUpdate{
				Value:  &dynamodb.AttributeValue{BOOL: aws.Bool(false)},
				Action: aws.String("PUT"),
			},
		},
	})
	return err
}

// GetLatestVersion gets the latest version of the secret from DynamoDB
func (d *dao) GetLatestVersion(secretName string) (*SecretRecord, error) {
	// Construct a query to the effect of:
	// Query the application secret table for a maximum of 1 item that has
	// 'Name' == secretName in the reverse order of the index. Since
	// records are incrementally indexed, we get the latest version back.
	expression := "#N = :val"
	expressionNames := map[string]string{"#N": "Name"}
	expressionValues := map[string]*dynamodb.AttributeValue{
		":val": &dynamodb.AttributeValue{
			S: aws.String(secretName),
		},
	}
	result, err := d.dynamodbClient.Query(&dynamodb.QueryInput{
		TableName:        aws.String(cfnclient.GetSecretsTableName(d.appName)),
		ScanIndexForward: aws.Bool(false),
		Limit:            aws.Int64(1),
		KeyConditionExpression:    aws.String(expression),
		ExpressionAttributeNames:  aws.StringMap(expressionNames),
		ExpressionAttributeValues: expressionValues,
	})

	if err != nil {
		return nil, err
	}

	if result == nil || aws.Int64Value(result.Count) == int64(0) {
		return nil, nil
	}

	loadedSecret := &SecretRecord{}
	err = dynamodbattribute.ConvertFromMap(result.Items[0], loadedSecret)
	return loadedSecret, err
}
