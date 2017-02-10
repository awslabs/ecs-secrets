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

package cmd

import (
	"flag"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/kms"
	cfnclient "github.com/awslabs/ecs-secrets/modules/cloudformation/client"
	mockcfnclient "github.com/awslabs/ecs-secrets/modules/cloudformation/client/mock"
	mockkmsclient "github.com/awslabs/ecs-secrets/modules/kms/client/mock"
	"github.com/golang/mock/gomock"
	"github.com/urfave/cli"
)

type alreadyExistsError struct{}

func (a alreadyExistsError) Error() string {
	return "already exists"
}

func (a alreadyExistsError) Code() string {
	return "AlreadyExistsException"
}

func (a alreadyExistsError) Message() string {
	return "AlreadyExistsException"
}

func (a alreadyExistsError) OrigErr() error {
	return fmt.Errorf("already exists")
}

func TestAliasAlreadyExistsErrorAWSError(t *testing.T) {
	if !aliasAlreadyExistsError(alreadyExistsError{}) {
		t.Error("alreadyExistsError not recognized as awserr.Error")
	}
}

func TestAliasAlreadyExistsErrorNonAWSError(t *testing.T) {
	if aliasAlreadyExistsError(fmt.Errorf("definitely not aws related")) {
		t.Error("Non awserr.Error recognized as aws eror")
	}
}

func TestDoSetupApplicationNameNotSet(t *testing.T) {
	flagSet := flag.NewFlagSet("ecs-secrets", 0)
	flagSet.String(fetchSecretsRoleFlag, "fetch", "")
	flagSet.String(createSecretsPrincipalFlag, "create", "")
	context := cli.NewContext(nil, flagSet, nil)
	err := doSetup(context, nil, nil)
	if err == nil {
		t.Error("Expected error when application name is not specified")
	}
}

func TestDoSetupCreateSecretsPrincipalNotSet(t *testing.T) {
	flagSet := flag.NewFlagSet("ecs-secrets", 0)
	flagSet.String(applicationNameFlag, "myapp", "")
	flagSet.String(fetchSecretsRoleFlag, "fetch", "")
	context := cli.NewContext(nil, flagSet, nil)
	err := doSetup(context, nil, nil)
	if err == nil {
		t.Error("Expected error when application name is not specified")
	}
}

func TestDoSetupFetchSecretsRoleNotSet(t *testing.T) {
	flagSet := flag.NewFlagSet("ecs-secrets", 0)
	flagSet.String(applicationNameFlag, "myapp", "")
	flagSet.String(createSecretsPrincipalFlag, "create", "")
	context := cli.NewContext(nil, flagSet, nil)
	err := doSetup(context, nil, nil)
	if err == nil {
		t.Error("Expected error when application name is not specified")
	}
}

func TestDoSetup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	flagSet := flag.NewFlagSet("ecs-secrets", 0)
	flagSet.String(fetchSecretsRoleFlag, "fetch", "")
	flagSet.String(createSecretsPrincipalFlag, "create", "")
	flagSet.String(applicationNameFlag, "myapp", "")
	context := cli.NewContext(nil, flagSet, nil)

	stacker := mockcfnclient.NewMockStacker(ctrl)
	kmsClient := mockkmsclient.NewMockClient(ctrl)

	mockCfnStackOutputs := []*cloudformation.Output{
		&cloudformation.Output{
			OutputKey:   aws.String(cfnclient.OutputMasterKey),
			OutputValue: aws.String("key-id"),
		},
		&cloudformation.Output{
			OutputKey:   aws.String(cfnclient.OutputSecretsDynamoTable),
			OutputValue: aws.String("secretsTable"),
		},
	}

	mockCfnStack := &cloudformation.Stack{
		Outputs: mockCfnStackOutputs,
	}
	gomock.InOrder(
		stacker.EXPECT().CreateStack("myapp", "create", "fetch").Return(mockCfnStack, nil),
		kmsClient.EXPECT().CreateAlias(&kms.CreateAliasInput{
			AliasName:   aws.String("alias/ECSSecretsMaskerKey-myapp"),
			TargetKeyId: aws.String("key-id"),
		}).Return(nil, nil),
	)
	err := doSetup(context, stacker, kmsClient)
	if err != nil {
		t.Errorf("Error setting up: %v", err)
	}
}

func TestDoSetupCMKAliasAlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	flagSet := flag.NewFlagSet("ecs-secrets", 0)
	flagSet.String(fetchSecretsRoleFlag, "fetch", "")
	flagSet.String(createSecretsPrincipalFlag, "create", "")
	flagSet.String(applicationNameFlag, "myapp", "")
	context := cli.NewContext(nil, flagSet, nil)

	stacker := mockcfnclient.NewMockStacker(ctrl)
	kmsClient := mockkmsclient.NewMockClient(ctrl)

	mockCfnStackOutputs := []*cloudformation.Output{
		&cloudformation.Output{
			OutputKey:   aws.String(cfnclient.OutputMasterKey),
			OutputValue: aws.String("key-id"),
		},
		&cloudformation.Output{
			OutputKey:   aws.String(cfnclient.OutputSecretsDynamoTable),
			OutputValue: aws.String("secretsTable"),
		},
	}

	mockCfnStack := &cloudformation.Stack{
		Outputs: mockCfnStackOutputs,
	}
	gomock.InOrder(
		stacker.EXPECT().CreateStack("myapp", "create", "fetch").Return(mockCfnStack, nil),
		kmsClient.EXPECT().CreateAlias(&kms.CreateAliasInput{
			AliasName:   aws.String("alias/ECSSecretsMaskerKey-myapp"),
			TargetKeyId: aws.String("key-id"),
		}).Return(nil, alreadyExistsError{}),
	)
	err := doSetup(context, stacker, kmsClient)
	if err != nil {
		t.Errorf("Error setting up: %v", err)
	}
}

func TestDoSetupCreateStackError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	flagSet := flag.NewFlagSet("ecs-secrets", 0)
	flagSet.String(fetchSecretsRoleFlag, "fetch", "")
	flagSet.String(createSecretsPrincipalFlag, "create", "")
	flagSet.String(applicationNameFlag, "myapp", "")
	context := cli.NewContext(nil, flagSet, nil)

	stacker := mockcfnclient.NewMockStacker(ctrl)

	stacker.EXPECT().CreateStack("myapp", "create", "fetch").Return(nil, fmt.Errorf("go away"))
	err := doSetup(context, stacker, nil)
	if err == nil {
		t.Errorf("Expected error setting up")
	}
}

func TestDoSetupCMKAliasError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	flagSet := flag.NewFlagSet("ecs-secrets", 0)
	flagSet.String(fetchSecretsRoleFlag, "fetch", "")
	flagSet.String(createSecretsPrincipalFlag, "create", "")
	flagSet.String(applicationNameFlag, "myapp", "")
	context := cli.NewContext(nil, flagSet, nil)

	stacker := mockcfnclient.NewMockStacker(ctrl)
	kmsClient := mockkmsclient.NewMockClient(ctrl)

	mockCfnStackOutputs := []*cloudformation.Output{
		&cloudformation.Output{
			OutputKey:   aws.String(cfnclient.OutputMasterKey),
			OutputValue: aws.String("key-id"),
		},
		&cloudformation.Output{
			OutputKey:   aws.String(cfnclient.OutputSecretsDynamoTable),
			OutputValue: aws.String("secretsTable"),
		},
	}

	mockCfnStack := &cloudformation.Stack{
		Outputs: mockCfnStackOutputs,
	}
	gomock.InOrder(
		stacker.EXPECT().CreateStack("myapp", "create", "fetch").Return(mockCfnStack, nil),
		kmsClient.EXPECT().CreateAlias(&kms.CreateAliasInput{
			AliasName:   aws.String("alias/ECSSecretsMaskerKey-myapp"),
			TargetKeyId: aws.String("key-id"),
		}).Return(nil, fmt.Errorf("why do you care?")),
	)
	err := doSetup(context, stacker, kmsClient)
	if err == nil {
		t.Errorf("Expected error setting up")
	}
}
