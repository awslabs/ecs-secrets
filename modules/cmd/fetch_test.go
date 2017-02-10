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
	"flag"
	"fmt"
	"testing"

	"github.com/awslabs/ecs-secrets/modules/api"

	"github.com/awslabs/ecs-secrets/modules/store/mock"
	"github.com/golang/mock/gomock"
	"github.com/urfave/cli"
)

func TestFetchCommandApplicationNameNotSet(t *testing.T) {
	flagSet := flag.NewFlagSet("ecs-secrets", 0)
	flagSet.String(nameFlag, "foo", "")
	flagSet.String(serialFlag, "1", "")
	context := cli.NewContext(nil, flagSet, nil)
	err := fetchCommand(context)
	if err == nil {
		t.Error("Expected error when application name is not specified")
	}
}

func TestDoFetchSecretNameNotSet(t *testing.T) {
	flagSet := flag.NewFlagSet("ecs-secrets", 0)
	flagSet.String(applicationNameFlag, "myapp", "")
	flagSet.String(serialFlag, "1", "")
	context := cli.NewContext(nil, flagSet, nil)
	err := doFetch(context, nil)
	if err == nil {
		t.Error("Expected error when name is not specified for the secret")
	}
}

func TestDoFetch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	flagSet := flag.NewFlagSet("ecs-secrets", 0)
	flagSet.String(applicationNameFlag, "myapp", "")
	flagSet.String(nameFlag, "name", "")
	context := cli.NewContext(nil, flagSet, nil)

	secretStore := mock_store.NewMockStore(ctrl)
	apiSecret := &api.SecretRecord{
		Name:    "name",
		Serial:  int64(1),
		Payload: "value",
		Active:  true,
	}
	secretStore.EXPECT().Get("name", "").Return(apiSecret, nil)
	err := doFetch(context, secretStore)
	if err != nil {
		t.Errorf("Error fetching secret: %v", err)
	}
}

func TestDoFetchOnSaveError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	flagSet := flag.NewFlagSet("ecs-secrets", 0)
	flagSet.String(applicationNameFlag, "myapp", "")
	flagSet.String(nameFlag, "name", "")
	context := cli.NewContext(nil, flagSet, nil)

	secretStore := mock_store.NewMockStore(ctrl)
	secretStore.EXPECT().Get("name", "").Return(nil, fmt.Errorf("ask me again tomorrow"))
	err := doFetch(context, secretStore)
	if err == nil {
		t.Error("Expected error fetching secret")
	}
}
