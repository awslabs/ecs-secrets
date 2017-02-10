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

import "github.com/aws/aws-sdk-go/service/kms"

//go:generate mockgen.sh github.com/awslabs/ecs-secrets/modules/kms/client Client mock/client_mock.go

// Client defines a subset of the kms client methods. The methods defined
// here are used to interact with KMS
type Client interface {
	CreateAlias(input *kms.CreateAliasInput) (*kms.CreateAliasOutput, error)
	Decrypt(*kms.DecryptInput) (*kms.DecryptOutput, error)
	GenerateDataKey(*kms.GenerateDataKeyInput) (*kms.GenerateDataKeyOutput, error)
}
