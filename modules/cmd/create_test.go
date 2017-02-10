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

// mockReader implements a mock file reader
type mockReader struct {
	payload []byte
	err     error
}

func (reader *mockReader) ReadFile(filename string) ([]byte, error) {
	return reader.payload, reader.err
}

func TestCreateCommandApplicationNameNotSet(t *testing.T) {
	flagSet := flag.NewFlagSet("ecs-secrets", 0)
	flagSet.String(nameFlag, "foo", "")
	flagSet.String(payloadFlag, "value", "")
	context := cli.NewContext(nil, flagSet, nil)
	err := createCommand(context)
	if err == nil {
		t.Error("Expected error when application name is not specified")
	}
}

func TestDoCreateSecretNameNotSet(t *testing.T) {
	flagSet := flag.NewFlagSet("ecs-secrets", 0)
	flagSet.String(applicationNameFlag, "myapp", "")
	flagSet.String(payloadFlag, "value", "")
	context := cli.NewContext(nil, flagSet, nil)
	err := doCreate(context, nil, nil)
	if err == nil {
		t.Error("Expected error when name is not specified for the secret")
	}
}

func TestDoCreateSecretPayloadAndLocationNotSet(t *testing.T) {
	flagSet := flag.NewFlagSet("ecs-secrets", 0)
	flagSet.String(applicationNameFlag, "myapp", "")
	flagSet.String(nameFlag, "name", "")
	context := cli.NewContext(nil, flagSet, nil)
	err := doCreate(context, nil, nil)
	if err == nil {
		t.Error("Expected error when value is not specified for the secret")
	}
}

func TestDoCreateSecretPayloadAndLocationSet(t *testing.T) {
	flagSet := flag.NewFlagSet("ecs-secrets", 0)
	flagSet.String(applicationNameFlag, "myapp", "")
	flagSet.String(nameFlag, "name", "")
	flagSet.String(payloadFlag, "value", "")
	flagSet.String(payloadLocationFlag, "value", "")
	context := cli.NewContext(nil, flagSet, nil)
	err := doCreate(context, nil, nil)
	if err == nil {
		t.Error("Expected error when value is not specified for the secret")
	}
}

func TestDoCreatePayloadSet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	flagSet := flag.NewFlagSet("ecs-secrets", 0)
	flagSet.String(applicationNameFlag, "myapp", "")
	flagSet.String(nameFlag, "name", "")
	flagSet.String(payloadFlag, "value", "")
	context := cli.NewContext(nil, flagSet, nil)

	secretStore := mock_store.NewMockStore(ctrl)
	apiSecret := &api.SecretRecord{
		Name:    "name",
		Serial:  int64(1),
		Payload: "value",
		Active:  true,
	}
	secretStore.EXPECT().Save(apiSecret).Return(apiSecret, nil)
	err := doCreate(context, secretStore, &mockReader{nil, nil})
	if err != nil {
		t.Errorf("Error creating secret: %v", err)
	}
}

func TestDoCreatePayloadLocationSet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	flagSet := flag.NewFlagSet("ecs-secrets", 0)
	flagSet.String(applicationNameFlag, "myapp", "")
	flagSet.String(nameFlag, "name", "")
	flagSet.String(payloadLocationFlag, "value", "")
	context := cli.NewContext(nil, flagSet, nil)

	secretStore := mock_store.NewMockStore(ctrl)
	apiSecret := &api.SecretRecord{
		Name:    "name",
		Serial:  int64(1),
		Payload: "value",
		Active:  true,
	}
	secretStore.EXPECT().Save(apiSecret).Return(apiSecret, nil)
	err := doCreate(context, secretStore, &mockReader{[]byte("value"), nil})
	if err != nil {
		t.Errorf("Error creating secret: %v", err)
	}
}

func TestDoCreateSaveError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	flagSet := flag.NewFlagSet("ecs-secrets", 0)
	flagSet.String(applicationNameFlag, "myapp", "")
	flagSet.String(nameFlag, "name", "")
	flagSet.String(payloadFlag, "value", "")
	context := cli.NewContext(nil, flagSet, nil)

	secretStore := mock_store.NewMockStore(ctrl)
	apiSecret := &api.SecretRecord{
		Name:    "name",
		Serial:  int64(1),
		Payload: "value",
		Active:  true,
	}
	secretStore.EXPECT().Save(apiSecret).Return(nil, fmt.Errorf("its beyond saving"))
	err := doCreate(context, secretStore, &mockReader{nil, nil})
	if err == nil {
		t.Errorf("Expected error creating secret")
	}
}
