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
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/awslabs/ecs-secrets/modules/dynamodb/client/mock"
	"github.com/golang/mock/gomock"
)

func TestGetSecretRecord(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ddbClient := mock_client.NewMockClient(ctrl)

	ddbClient.EXPECT().GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("ECS-Secrets-myapp-Secrets"),
		Key: map[string]*dynamodb.AttributeValue{
			"Name": {
				S: aws.String("foo"),
			},
			"Serial": {
				N: aws.String("1"),
			},
		},
	}).Return(&dynamodb.GetItemOutput{
		Item: map[string]*dynamodb.AttributeValue{
			"Name": &dynamodb.AttributeValue{
				S: aws.String("foo-name"),
			},
		},
	}, nil)

	dao := NewDAO("myapp", ddbClient)
	secret, err := dao.GetSecretRecord("foo", 1)
	if err != nil {
		t.Fatalf("Error getting secret record: %v", err)
	}
	expectedSecret := &SecretRecord{
		Name: "foo-name",
	}
	if !reflect.DeepEqual(secret, expectedSecret) {
		t.Errorf("Mismatch between expected and recieved secret: %v != %v", secret, expectedSecret)
	}
}

func TestGetSecretRecordGetItemReturnsEmptyItemMap(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ddbClient := mock_client.NewMockClient(ctrl)

	ddbClient.EXPECT().GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("ECS-Secrets-myapp-Secrets"),
		Key: map[string]*dynamodb.AttributeValue{
			"Name": {
				S: aws.String("foo"),
			},
			"Serial": {
				N: aws.String("1"),
			},
		},
	}).Return(&dynamodb.GetItemOutput{
		Item: map[string]*dynamodb.AttributeValue{},
	}, nil)

	dao := NewDAO("myapp", ddbClient)
	_, err := dao.GetSecretRecord("foo", 1)
	if err == nil {
		t.Error("Expected error getting secret record")
	}
}

func TestGetSecretRecordGetItemReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ddbClient := mock_client.NewMockClient(ctrl)

	ddbClient.EXPECT().GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("ECS-Secrets-myapp-Secrets"),
		Key: map[string]*dynamodb.AttributeValue{
			"Name": {
				S: aws.String("foo"),
			},
			"Serial": {
				N: aws.String("1"),
			},
		},
	}).Return(nil, fmt.Errorf("no item"))

	dao := NewDAO("myapp", ddbClient)
	_, err := dao.GetSecretRecord("foo", 1)
	if err == nil {
		t.Error("Expected error getting secret record")
	}
}

func TestPutSecretRecord(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ddbClient := mock_client.NewMockClient(ctrl)

	ddbClient.EXPECT().PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("ECS-Secrets-myapp-Secrets"),
		Item: map[string]*dynamodb.AttributeValue{
			"Name": {
				S: aws.String("foo-name"),
			},
			"Serial": {
				N: aws.String("1"),
			},
			"EncryptedData": {
				S: aws.String("foo-data"),
			},
			"EncryptedDataKey": {
				S: aws.String("foo-data-key"),
			},
			"Active": {
				BOOL: aws.Bool(true),
			},
		},
	}).Return(nil, nil)
	dao := NewDAO("myapp", ddbClient)
	secret := &SecretRecord{
		Name:             "foo-name",
		Serial:           1,
		EncryptedData:    "foo-data",
		EncryptedDataKey: "foo-data-key",
		Active:           true,
	}
	err := dao.PutSecretRecord(secret)
	if err != nil {
		t.Errorf("Error putting secret record: %v", err)
	}
}

func TestPutSecretRecordError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ddbClient := mock_client.NewMockClient(ctrl)

	ddbClient.EXPECT().PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("ECS-Secrets-myapp-Secrets"),
		Item: map[string]*dynamodb.AttributeValue{
			"Name": {
				S: aws.String("foo-name"),
			},
			"Serial": {
				N: aws.String("1"),
			},
			"EncryptedData": {
				S: aws.String("foo-data"),
			},
			"EncryptedDataKey": {
				S: aws.String("foo-data-key"),
			},
			"Active": {
				BOOL: aws.Bool(true),
			},
		},
	}).Return(nil, fmt.Errorf("i have secrets of my own, you know?"))
	dao := NewDAO("myapp", ddbClient)
	secret := &SecretRecord{
		Name:             "foo-name",
		Serial:           1,
		EncryptedData:    "foo-data",
		EncryptedDataKey: "foo-data-key",
		Active:           true,
	}
	err := dao.PutSecretRecord(secret)
	if err == nil {
		t.Error("Expected error putting secret record")
	}
}

func TestRevokeSecretRecord(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ddbClient := mock_client.NewMockClient(ctrl)

	ddbClient.EXPECT().UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String("ECS-Secrets-myapp-Secrets"),
		Key: map[string]*dynamodb.AttributeValue{
			"Name": {
				S: aws.String("foo"),
			},
			"Serial": {
				N: aws.String("1"),
			},
		},
		AttributeUpdates: map[string]*dynamodb.AttributeValueUpdate{
			"Active": &dynamodb.AttributeValueUpdate{
				Value: &dynamodb.AttributeValue{
					BOOL: aws.Bool(false),
				},
				Action: aws.String("PUT"),
			},
		},
	}).Return(nil, nil)
	dao := NewDAO("myapp", ddbClient)
	err := dao.RevokeSecretRecord("foo", 1)
	if err != nil {
		t.Errorf("Error revoking secret record: %v", err)
	}
}

func TestGetLatestVersion(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ddbClient := mock_client.NewMockClient(ctrl)

	ddbClient.EXPECT().Query(&dynamodb.QueryInput{
		TableName:        aws.String("ECS-Secrets-myapp-Secrets"),
		ScanIndexForward: aws.Bool(false),
		Limit:            aws.Int64(1),
		KeyConditionExpression:   aws.String("#N = :val"),
		ExpressionAttributeNames: aws.StringMap(map[string]string{"#N": "Name"}),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":val": &dynamodb.AttributeValue{
				S: aws.String("foo"),
			},
		},
	}).Return(&dynamodb.QueryOutput{
		Count: aws.Int64(1),
		Items: []map[string]*dynamodb.AttributeValue{
			map[string]*dynamodb.AttributeValue{
				"Name": {
					S: aws.String("foo-name"),
				},
			},
		},
	}, nil)
	dao := NewDAO("myapp", ddbClient)
	secret, err := dao.GetLatestVersion("foo")
	if err != nil {
		t.Fatalf("Error getting latest version: %v", err)
	}
	expectedSecret := &SecretRecord{
		Name: "foo-name",
	}
	if !reflect.DeepEqual(secret, expectedSecret) {
		t.Errorf("Mismatch between expected and recieved secret: %v != %v", secret, expectedSecret)
	}
}

func TestGetLatestVersionReturnsEmptyItems(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ddbClient := mock_client.NewMockClient(ctrl)

	ddbClient.EXPECT().Query(&dynamodb.QueryInput{
		TableName:        aws.String("ECS-Secrets-myapp-Secrets"),
		ScanIndexForward: aws.Bool(false),
		Limit:            aws.Int64(1),
		KeyConditionExpression:   aws.String("#N = :val"),
		ExpressionAttributeNames: aws.StringMap(map[string]string{"#N": "Name"}),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":val": &dynamodb.AttributeValue{
				S: aws.String("foo"),
			},
		},
	}).Return(&dynamodb.QueryOutput{
		Count: aws.Int64(0),
	}, nil)
	dao := NewDAO("myapp", ddbClient)
	secret, err := dao.GetLatestVersion("foo")
	if err != nil {
		t.Fatalf("Error getting latest version: %v", err)
	}

	if secret != nil {
		t.Error("Expected empty secret record")
	}
}

func TestGetLatestVersionQueryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ddbClient := mock_client.NewMockClient(ctrl)

	ddbClient.EXPECT().Query(&dynamodb.QueryInput{
		TableName:        aws.String("ECS-Secrets-myapp-Secrets"),
		ScanIndexForward: aws.Bool(false),
		Limit:            aws.Int64(1),
		KeyConditionExpression:   aws.String("#N = :val"),
		ExpressionAttributeNames: aws.StringMap(map[string]string{"#N": "Name"}),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":val": &dynamodb.AttributeValue{
				S: aws.String("foo"),
			},
		},
	}).Return(nil, fmt.Errorf("enough already"))
	dao := NewDAO("myapp", ddbClient)
	_, err := dao.GetLatestVersion("foo")
	if err == nil {
		t.Errorf("Expected error getting latest version")
	}
}
