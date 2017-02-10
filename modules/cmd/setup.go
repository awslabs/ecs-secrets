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

package cmd

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/kms"
	cfnclient "github.com/awslabs/ecs-secrets/modules/cloudformation/client"
	kmsclient "github.com/awslabs/ecs-secrets/modules/kms/client"
	"github.com/awslabs/ecs-secrets/modules/kms/utils"
	log "github.com/cihub/seelog"
	"github.com/urfave/cli"
)

const (
	secretsTableQueryPolicyStatement = `{
    "Effect": "Allow",
    "Action": [
	"dynamodb:Query",
	"dynamodb:GetItem"
    ],
    "Resource": [
	"%s"
    ]
}`

	secretsTablePutPolicyStatement = `{
    "Effect": "Allow",
    "Action": [
	"dynamodb:PutItem",
	"dynamodb:UpdateItem"
    ],
    "Resource": [
	"%s"
    ]
}`
)

func setupCommand(context *cli.Context) error {
	stacker := cfnclient.NewStacker(cloudformation.New(session.New()))
	return doSetup(context, stacker, kms.New(session.New()))
}

func doSetup(context *cli.Context, stacker cfnclient.Stacker, kmsClient kmsclient.Client) error {
	// Validate that application name has been specified
	appName, err := getRequiredArgumentFromFlag(context, applicationNameFlag)
	if err != nil {
		return err
	}

	// Validate that IAM principal for creating secrets has been specified
	createSecretsPrincipal, err := getRequiredArgumentFromFlag(context, createSecretsPrincipalFlag)
	if err != nil {
		return err
	}

	// Validate that IAM Role for fetching secrets has been specified
	fetchSecretsRole, err := getRequiredArgumentFromFlag(context, fetchSecretsRoleFlag)
	if err != nil {
		return err
	}

	log.Debugf("Setting up stack for application: %s", appName)
	stack, err := stacker.CreateStack(appName, createSecretsPrincipal, fetchSecretsRole)
	if err != nil {
		return fmt.Errorf("Error creating cloudformation stack: %v", err)
	}

	secretsTable, err := cfnclient.GetCreatedSecretsTableName(stack)
	if err != nil {
		return err
	}
	log.Infof("Secrets are stored in the table: %s", secretsTable)
	log.Infof("Update '%s' to provide read access for this table by updating the policy statement with: %s",
		fetchSecretsRole, fmt.Sprintf(secretsTableQueryPolicyStatement, secretsTable))
	log.Infof("Update '%s' to provide write access for this table by updating the policy statement with: %s",
		createSecretsPrincipal, fmt.Sprintf(secretsTablePutPolicyStatement, secretsTable))

	// Set the alias for the secret after stack creation completes
	return setCMKAlias(appName, stack, kmsClient)

}

func setCMKAlias(appName string, stack *cloudformation.Stack, kmsClient kmsclient.Client) error {
	// Get KMS CMK ID from cloudformation stack outputs
	createdKMSCMKID, err := cfnclient.GetCreatedKMSCMKID(stack)
	if err != nil {
		return fmt.Errorf("Error getting customer master key for application: %v", err)
	}

	// Try creating an alias for the master key. There's no list-aliases-for-key API in KMS.
	// So, blindly invoke create-alias and handle the 'already exists error' appropriately
	log.Debugf("Created kms key: %s", createdKMSCMKID)
	_, err = kmsClient.CreateAlias(&kms.CreateAliasInput{
		AliasName:   aws.String(utils.GetCMKAlias(appName)),
		TargetKeyId: aws.String(createdKMSCMKID),
	})
	if err != nil {
		if aliasAlreadyExistsError(err) {
			log.Debugf("Alias already exists for kms key: %s", createdKMSCMKID)
			log.Info("Setup complete")
			return nil
		}
		return fmt.Errorf("Error creating KMS alias for MasterKey: %v", err)

	}

	log.Debugf("Alias set for kms key: %s", createdKMSCMKID)
	log.Info("Setup complete")

	return nil
}

func aliasAlreadyExistsError(err error) bool {
	if awsErr, ok := err.(awserr.Error); ok {
		if awsErr.Code() == "AlreadyExistsException" {
			return true
		}
	}
	return false
}
