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

package client

func getTemplate() string {
	return template
}

// TODO: Parameterize ReadCapacityUnits and WriteCapacityUnits
var template = `
{
  "AWSTemplateFormatVersion": "2010-09-09",
  "Description" : "AWS CloudFormation template to create resources for ECS Secrets.",
  "Parameters": {
    "ECSSecretsTableName": {
      "Type": "String",
      "Description": "Table name for ECS Secrets"
    },
    "ECSSecretsIAMPrincipalForCreatingSecrets": {
      "Type": "String",
      "Description": "IAM Entity used to create secrets"
    },
    "ECSSecretsIAMRoleArn": {
      "Type": "String",
      "Description": "Task IAM Role Arn used by the application"
    }
  },
  "Resources" : {
    "ECSSecretsTable" : {
      "Type" : "AWS::DynamoDB::Table",
      "Properties" : {
        "AttributeDefinitions" : [
          {
            "AttributeName" : "Name",
            "AttributeType" : "S"
          },
          {
            "AttributeName" : "Serial",
            "AttributeType" : "N"
          }
        ],
        "KeySchema" : [
          {
            "AttributeName" : "Name",
            "KeyType" : "HASH"
          },
          {
            "AttributeName" : "Serial",
            "KeyType" : "RANGE"
          }
        ],
        "ProvisionedThroughput" : {
          "ReadCapacityUnits" : "5",
          "WriteCapacityUnits" : "5"
        },
        "TableName" : {"Ref": "ECSSecretsTableName"}
      }
    },
    "ECSSecretsMasterKey": {
      "Type" : "AWS::KMS::Key",
      "Properties" : {
        "Description" : "Master Key for ECS Secrets",
        "KeyPolicy" : {
          "Version": "2012-10-17",
          "Id": "ecs-secrets-setup-key-policy",
          "Statement": [
            {
              "Sid": "Allow administration of the key",
              "Effect": "Allow",
              "Principal": { 
                "AWS": { "Fn::Join": [":", ["arn:aws:iam:", { "Ref":"AWS::AccountId" }, "root"]]}
              },
              "Action": [
                "kms:Create*",
                "kms:Describe*",
                "kms:Enable*",
                "kms:List*",
                "kms:Put*",
                "kms:Update*",
                "kms:Revoke*",
                "kms:Disable*",
                "kms:Get*",
                "kms:Delete*",
                "kms:ScheduleKeyDeletion",
                "kms:CancelKeyDeletion"
              ],
              "Resource": "*"
            },
            {
              "Sid": "Allow use of the key to create secrets",
              "Effect": "Allow",
              "Principal": { "AWS": { "Ref": "ECSSecretsIAMPrincipalForCreatingSecrets" } },
              "Action": [
                "kms:Encrypt",
                "kms:Decrypt",
                "kms:ReEncrypt",
                "kms:GenerateDataKey*",
                "kms:DescribeKey"
              ], 
              "Resource": "*"
            },
            {
              "Sid": "Allow use of the key to retrieve secrets",
              "Effect": "Allow",
              "Principal": { "AWS": { "Ref": "ECSSecretsIAMRoleArn" } },
              "Action": [
                "kms:Decrypt",
                "kms:DescribeKey"
              ], 
              "Resource": "*"
            }
          ]
        }
      }
    }
  },
  "Outputs" : {
    "secretsDynamoTable" : {
      "Value" : { "Fn::Sub" : "arn:aws:dynamodb:${AWS::Region}:${AWS::AccountId}:table/${ECSSecretsTable}" }
    },
    "kmsKey": {
      "Value" : { "Ref" : "ECSSecretsMasterKey" }
    }
  }
}
`
