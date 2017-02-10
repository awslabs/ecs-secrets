// Copyright 2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package client

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/awslabs/ecs-secrets/modules/cloudformation/client/mock"
	"github.com/golang/mock/gomock"
)

func TestGetCreatedKMSCMKIDNotFound(t *testing.T) {
	stackOutputs := []*cloudformation.Output{}

	stack := &cloudformation.Stack{
		Outputs: stackOutputs,
	}

	_, err := GetCreatedKMSCMKID(stack)
	if err == nil {
		t.Error("Expected error finding output master key in stack")
	}
}

func TestGetCreatedKMSCMKID(t *testing.T) {
	stackOutputs := []*cloudformation.Output{
		&cloudformation.Output{
			OutputKey:   aws.String(OutputMasterKey),
			OutputValue: aws.String("key-id"),
		},
	}

	stack := &cloudformation.Stack{
		Outputs: stackOutputs,
	}

	keyID, err := GetCreatedKMSCMKID(stack)
	if err != nil {
		t.Errorf("Error finding output master key in stack: %v", err)
	}
	if keyID != "key-id" {
		t.Error("Unexpected key-id in stack output for master key")
	}
}

func TestGetSecretsTableName(t *testing.T) {
	if GetSecretsTableName("app") != "ECS-Secrets-app-Secrets" {
		t.Error("Mismtach in expected secrets table name")
	}
}

func TestGetStackName(t *testing.T) {
	if getStackName("app") != "ECS-Secrets-app" {
		t.Error("Mismtach in expected stack name")
	}
}

func TestGetCreatedSecretsTableName(t *testing.T) {
	stackOutputs := []*cloudformation.Output{
		&cloudformation.Output{
			OutputKey:   aws.String(OutputSecretsDynamoTable),
			OutputValue: aws.String("secret"),
		},
	}

	stack := &cloudformation.Stack{
		Outputs: stackOutputs,
	}

	secretsTable, err := GetCreatedSecretsTableName(stack)
	if err != nil {
		t.Errorf("Error finding output secrets table name: %v", err)
	}
	if secretsTable != "secret" {
		t.Error("Unexpected secrets table name in stack output")
	}
}

func TestCreatedGetSecretsTableNameNotFound(t *testing.T) {
	stackOutputs := []*cloudformation.Output{}

	stack := &cloudformation.Stack{
		Outputs: stackOutputs,
	}

	_, err := GetCreatedSecretsTableName(stack)
	if err == nil {
		t.Error("Expected error getting table name from stack")
	}
}

func TestCreateStackExistingStack(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock_client.NewMockClient(ctrl)
	mockClient.EXPECT().DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String("ECS-Secrets-myapp"),
	}).Return(&cloudformation.DescribeStacksOutput{
		Stacks: []*cloudformation.Stack{
			&cloudformation.Stack{},
		},
	}, nil)
	stacker := NewStacker(mockClient)
	_, err := stacker.CreateStack("myapp", "create", "fetch")
	if err != nil {
		t.Errorf("Error creating stack: %v", err)
	}
}

func TestCreateStackNewStack(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock_client.NewMockClient(ctrl)
	gomock.InOrder(
		mockClient.EXPECT().DescribeStacks(&cloudformation.DescribeStacksInput{
			StackName: aws.String("ECS-Secrets-myapp"),
		}).Return(nil, fmt.Errorf("stack is a queue now")),
		mockClient.EXPECT().CreateStack(&cloudformation.CreateStackInput{
			TemplateBody: aws.String(getTemplate()),
			StackName:    aws.String("ECS-Secrets-myapp"),
			Parameters: []*cloudformation.Parameter{
				&cloudformation.Parameter{
					ParameterKey:   aws.String("ECSSecretsTableName"),
					ParameterValue: aws.String("ECS-Secrets-myapp-Secrets"),
				},
				&cloudformation.Parameter{
					ParameterKey:   aws.String("ECSSecretsIAMPrincipalForCreatingSecrets"),
					ParameterValue: aws.String("create"),
				},
				&cloudformation.Parameter{
					ParameterKey:   aws.String("ECSSecretsIAMRoleArn"),
					ParameterValue: aws.String("fetch"),
				},
			},
		}).Return(nil, nil),
		mockClient.EXPECT().WaitUntilStackCreateComplete(&cloudformation.DescribeStacksInput{
			StackName: aws.String("ECS-Secrets-myapp"),
		}).Return(nil),
		mockClient.EXPECT().DescribeStacks(&cloudformation.DescribeStacksInput{
			StackName: aws.String("ECS-Secrets-myapp"),
		}).Return(&cloudformation.DescribeStacksOutput{
			Stacks: []*cloudformation.Stack{
				&cloudformation.Stack{},
			},
		}, nil),
	)
	stacker := NewStacker(mockClient)
	_, err := stacker.CreateStack("myapp", "create", "fetch")
	if err != nil {
		t.Errorf("Error creating stack: %v", err)
	}
}

func TestCreateStackNewStackOnCreateStackError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock_client.NewMockClient(ctrl)
	gomock.InOrder(
		mockClient.EXPECT().DescribeStacks(&cloudformation.DescribeStacksInput{
			StackName: aws.String("ECS-Secrets-myapp"),
		}).Return(nil, fmt.Errorf("stack is a queue now")),
		mockClient.EXPECT().CreateStack(&cloudformation.CreateStackInput{
			TemplateBody: aws.String(getTemplate()),
			StackName:    aws.String("ECS-Secrets-myapp"),
			Parameters: []*cloudformation.Parameter{
				&cloudformation.Parameter{
					ParameterKey:   aws.String("ECSSecretsTableName"),
					ParameterValue: aws.String("ECS-Secrets-myapp-Secrets"),
				},
				&cloudformation.Parameter{
					ParameterKey:   aws.String("ECSSecretsIAMPrincipalForCreatingSecrets"),
					ParameterValue: aws.String("create"),
				},
				&cloudformation.Parameter{
					ParameterKey:   aws.String("ECSSecretsIAMRoleArn"),
					ParameterValue: aws.String("fetch"),
				},
			},
		}).Return(nil, fmt.Errorf("one cannot always get what one wants")),
	)
	stacker := NewStacker(mockClient)
	_, err := stacker.CreateStack("myapp", "create", "fetch")
	if err == nil {
		t.Error("Expected error creating stack")
	}
}

func TestCreateStackNewStackOnWaitError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock_client.NewMockClient(ctrl)
	gomock.InOrder(
		mockClient.EXPECT().DescribeStacks(&cloudformation.DescribeStacksInput{
			StackName: aws.String("ECS-Secrets-myapp"),
		}).Return(nil, fmt.Errorf("stack is a queue now")),
		mockClient.EXPECT().CreateStack(&cloudformation.CreateStackInput{
			TemplateBody: aws.String(getTemplate()),
			StackName:    aws.String("ECS-Secrets-myapp"),
			Parameters: []*cloudformation.Parameter{
				&cloudformation.Parameter{
					ParameterKey:   aws.String("ECSSecretsTableName"),
					ParameterValue: aws.String("ECS-Secrets-myapp-Secrets"),
				},
				&cloudformation.Parameter{
					ParameterKey:   aws.String("ECSSecretsIAMPrincipalForCreatingSecrets"),
					ParameterValue: aws.String("create"),
				},
				&cloudformation.Parameter{
					ParameterKey:   aws.String("ECSSecretsIAMRoleArn"),
					ParameterValue: aws.String("fetch"),
				},
			},
		}).Return(nil, nil),
		mockClient.EXPECT().WaitUntilStackCreateComplete(&cloudformation.DescribeStacksInput{
			StackName: aws.String("ECS-Secrets-myapp"),
		}).Return(fmt.Errorf("sorry, afk")),
	)
	stacker := NewStacker(mockClient)
	_, err := stacker.CreateStack("myapp", "create", "fetch")
	if err == nil {
		t.Error("Expected error creating stack")
	}
}

func TestDeleteStack(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock_client.NewMockClient(ctrl)
	gomock.InOrder(
		mockClient.EXPECT().DescribeStacks(&cloudformation.DescribeStacksInput{
			StackName: aws.String("ECS-Secrets-myapp"),
		}).Return(&cloudformation.DescribeStacksOutput{
			Stacks: []*cloudformation.Stack{
				&cloudformation.Stack{},
			},
		}, nil),
		mockClient.EXPECT().DeleteStack(&cloudformation.DeleteStackInput{
			StackName: aws.String("ECS-Secrets-myapp"),
		}).Return(nil, nil),
		mockClient.EXPECT().WaitUntilStackDeleteComplete(&cloudformation.DescribeStacksInput{
			StackName: aws.String("ECS-Secrets-myapp"),
		}).Return(nil),
	)
	stacker := NewStacker(mockClient)
	err := stacker.DeleteStack("myapp")
	if err != nil {
		t.Errorf("Error deleting stack: %v", err)
	}
}

func TestDeleteStackOnDescribeError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock_client.NewMockClient(ctrl)
	gomock.InOrder(
		mockClient.EXPECT().DescribeStacks(&cloudformation.DescribeStacksInput{
			StackName: aws.String("ECS-Secrets-myapp"),
		}).Return(nil, fmt.Errorf("stack? what stack?")),
	)
	stacker := NewStacker(mockClient)
	err := stacker.DeleteStack("myapp")
	if err == nil {
		t.Error("Expected error deleting stack")
	}
}

func TestDeleteStackOnDeleteStackError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock_client.NewMockClient(ctrl)
	gomock.InOrder(
		mockClient.EXPECT().DescribeStacks(&cloudformation.DescribeStacksInput{
			StackName: aws.String("ECS-Secrets-myapp"),
		}).Return(&cloudformation.DescribeStacksOutput{
			Stacks: []*cloudformation.Stack{
				&cloudformation.Stack{},
			},
		}, nil),
		mockClient.EXPECT().DeleteStack(&cloudformation.DeleteStackInput{
			StackName: aws.String("ECS-Secrets-myapp"),
		}).Return(nil, fmt.Errorf("tough luck")),
	)
	stacker := NewStacker(mockClient)
	err := stacker.DeleteStack("myapp")
	if err == nil {
		t.Error("Expected error deleting stack")
	}
}

func TestDeleteStackOnWaitError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock_client.NewMockClient(ctrl)
	gomock.InOrder(
		mockClient.EXPECT().DescribeStacks(&cloudformation.DescribeStacksInput{
			StackName: aws.String("ECS-Secrets-myapp"),
		}).Return(&cloudformation.DescribeStacksOutput{
			Stacks: []*cloudformation.Stack{
				&cloudformation.Stack{},
			},
		}, nil),
		mockClient.EXPECT().DeleteStack(&cloudformation.DeleteStackInput{
			StackName: aws.String("ECS-Secrets-myapp"),
		}).Return(nil, nil),
		mockClient.EXPECT().WaitUntilStackDeleteComplete(&cloudformation.DescribeStacksInput{
			StackName: aws.String("ECS-Secrets-myapp"),
		}).Return(fmt.Errorf("vacationing")),
	)
	stacker := NewStacker(mockClient)
	err := stacker.DeleteStack("myapp")
	if err == nil {
		t.Error("Expected error deleting stack")
	}
}
