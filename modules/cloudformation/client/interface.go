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

import "github.com/aws/aws-sdk-go/service/cloudformation"

//go:generate mockgen.sh github.com/awslabs/ecs-secrets/modules/cloudformation/client Client mock/client_mock.go

// Client defines a subset of the cloudformation client methods. The methods defined
// here are used to interact with Cloudformation service
type Client interface {
	CreateStack(*cloudformation.CreateStackInput) (*cloudformation.CreateStackOutput, error)
	DeleteStack(*cloudformation.DeleteStackInput) (*cloudformation.DeleteStackOutput, error)
	DescribeStacks(*cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error)
	DescribeStackEvents(input *cloudformation.DescribeStackEventsInput) (*cloudformation.DescribeStackEventsOutput, error)
	WaitUntilStackCreateComplete(input *cloudformation.DescribeStacksInput) error
	WaitUntilStackDeleteComplete(input *cloudformation.DescribeStacksInput) error
}
