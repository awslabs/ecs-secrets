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
	"io/ioutil"

	"github.com/awslabs/ecs-secrets/modules/api"
	"github.com/awslabs/ecs-secrets/modules/store"
	log "github.com/cihub/seelog"
	"github.com/urfave/cli"
)

// fileReader interface wraps the ioutil.ReadFile method. This makes it easier
// to mock and test the read-secret-payload-from-file functionality.
type fileReader interface {
	ReadFile(filename string) ([]byte, error)
}

type ioutilFileReader struct{}

func createCommand(context *cli.Context) error {
	// Validate that application name has been specified
	appName, err := getRequiredArgumentFromFlag(context, applicationNameFlag)
	if err != nil {
		return err
	}
	return doCreate(context, createSecretStore(appName), &ioutilFileReader{})
}

func doCreate(context *cli.Context, secretStore store.Store, reader fileReader) error {
	// Validate that secrets name has been specified
	name, err := getRequiredArgumentFromFlag(context, nameFlag)
	if err != nil {
		return err
	}

	// Validate that either payload or payload location are specifiec
	payload, payloadErr := getRequiredArgumentFromFlag(context, payloadFlag)
	payloadLocation, payloadLocationErr := getRequiredArgumentFromFlag(context, payloadLocationFlag)

	if payloadErr == nil && payloadLocationErr == nil {
		return fmt.Errorf("Incorrect usage. Only one of '%s' or '%s' should be specified", payloadFlag, payloadLocationFlag)
	}

	if payloadErr != nil {
		if payloadLocationErr != nil {
			return fmt.Errorf("Incorrect usage. One of '%s' or '%s' should be specified", payloadFlag, payloadLocationFlag)
		}
		log.Debugf("Reading from %s", payloadLocation)
		readBytes, err := reader.ReadFile(payloadLocation)
		if err != nil {
			return fmt.Errorf("Error reading from %s: %v", payloadLocation, err)
		}
		payload = string(readBytes)
	}

	log.Debugf("Creating secret with name: %s", name)
	editedSecret, err := secretStore.Save(&api.SecretRecord{
		Name:    name,
		Serial:  int64(1),
		Payload: payload,
		Active:  true,
	})
	if err != nil {
		return err
	}

	log.Infof("Created application secret with name: %s and version: %d", editedSecret.Name, editedSecret.Serial)
	return nil
}

func (reader *ioutilFileReader) ReadFile(filename string) ([]byte, error) {
	return ioutil.ReadFile(filename)
}
