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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	log "github.com/cihub/seelog"
)

//go:generate mockgen.sh github.com/awslabs/ecs-secrets/modules/cloudformation/client Client mock/client_mock.go

const (
	OutputMasterKey          = "kmsKey"
	OutputSecretsDynamoTable = "secretsDynamoTable"

	stackPrefix            = "ECS-Secrets-"
	secretsTableNameSuffix = "-Secrets"

	parameterKeySecretsTableName = "ECSSecretsTableName"

	parameterKeyIAMPrincipalForCreatingSecrets = "ECSSecretsIAMPrincipalForCreatingSecrets"
	paramaterKeyIAMRoleArn                     = "ECSSecretsIAMRoleArn"
)

// Stacker defines an interface for creating the cloudformation stack required for
// saving and retrieving secrets
type Stacker interface {
	CreateStack(appName string, createSecretsPrincipal string, fetchSecretsRole string) (*cloudformation.Stack, error)
	DeleteStack(appName string) error
}

// stacker implements the Stacker interface
type stacker struct {
	client Client
}

// NewStacker returns a new 'stacker' object that implements the Stacker interface
func NewStacker(client Client) Stacker {
	return &stacker{
		client: client,
	}
}

// GetCreatedKMSCMKID returns the master key id for the created KMS Key from the
// cloudformation stack output
func GetCreatedKMSCMKID(stack *cloudformation.Stack) (string, error) {
	for _, output := range stack.Outputs {
		if aws.StringValue(output.OutputKey) == OutputMasterKey {
			return aws.StringValue(output.OutputValue), nil
		}
	}

	return "", fmt.Errorf("Unable to get customer master key id")
}

// GetSecretsTableName returns the dynamodb table created for storing secrets from
// the cloudformation stack output
func GetCreatedSecretsTableName(stack *cloudformation.Stack) (string, error) {
	for _, output := range stack.Outputs {
		if aws.StringValue(output.OutputKey) == OutputSecretsDynamoTable {
			return aws.StringValue(output.OutputValue), nil
		}
	}

	return "", fmt.Errorf("Unable to get secrets table name")
}

// GetSecretsTableName returns the DynamoDB table used for storing secrets
func GetSecretsTableName(appName string) string {
	return getStackName(appName) + secretsTableNameSuffix
}

// CreateStack creates the cloudformation stack
func (s *stacker) CreateStack(appName string, createSecretsPrincipal string, fetchSecretsRole string) (*cloudformation.Stack, error) {
	stackName := getStackName(appName)
	stack, err := s.describeStack(stackName)
	if err != nil {
		log.Infof("Unable to describe stack: %s, creating a new one", stackName)

		return s.createStack(appName, createSecretsPrincipal, fetchSecretsRole)
	}

	log.Infof("Returning existing stack: %s", stackName)
	return stack, nil
}

// DeleteStack deletes the cloudformation stack
func (s *stacker) DeleteStack(appName string) error {
	stackName := getStackName(appName)
	_, err := s.describeStack(stackName)
	if err != nil {
		return fmt.Errorf("Could not find stack for: %s", appName)
	}
	_, err = s.client.DeleteStack(&cloudformation.DeleteStackInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		return err
	}
	return s.client.WaitUntilStackDeleteComplete(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
}

func (s *stacker) describeStack(stackName string) (*cloudformation.Stack, error) {
	describeStacksResponse, err := s.client.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		return nil, err
	}
	if len(describeStacksResponse.Stacks) == 0 {
		return nil, fmt.Errorf("Unable to describe stack: %s", stackName)
	}

	return describeStacksResponse.Stacks[0], nil
}

func (s *stacker) createStack(appName string, createSecretsPrincipal string, fetchSecretsRole string) (*cloudformation.Stack, error) {
	stackName := getStackName(appName)

	params := []*cloudformation.Parameter{
		{
			ParameterKey:   aws.String(parameterKeySecretsTableName),
			ParameterValue: aws.String(GetSecretsTableName(appName)),
		},
		{
			ParameterKey:   aws.String(parameterKeyIAMPrincipalForCreatingSecrets),
			ParameterValue: aws.String(createSecretsPrincipal),
		},
		{
			ParameterKey:   aws.String(paramaterKeyIAMRoleArn),
			ParameterValue: aws.String(fetchSecretsRole),
		},
	}

	_, err := s.client.CreateStack(&cloudformation.CreateStackInput{
		TemplateBody: aws.String(getTemplate()),
		StackName:    aws.String(stackName),
		Parameters:   params,
	})

	if err != nil {
		return nil, err
	}

	err = s.client.WaitUntilStackCreateComplete(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})

	if err != nil {
		return nil, err
	}

	return s.describeStack(stackName)
}

func getStackName(appName string) string {
	return stackPrefix + appName
}
