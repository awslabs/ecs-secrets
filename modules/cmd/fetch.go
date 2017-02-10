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
	"encoding/json"
	"fmt"

	log "github.com/cihub/seelog"

	"github.com/awslabs/ecs-secrets/modules/store"
	"github.com/urfave/cli"
)

func fetchCommand(context *cli.Context) error {
	// Validate that application name has been specified
	appName, err := getRequiredArgumentFromFlag(context, applicationNameFlag)
	if err != nil {
		return err
	}
	return doFetch(context, createSecretStore(appName))
}

func doFetch(context *cli.Context, secretStore store.Store) error {
	// Validate secrets name has been specified
	name, err := getRequiredArgumentFromFlag(context, nameFlag)
	if err != nil {
		return err
	}

	serial := context.String(serialFlag)
	log.Debugf("Fetching secret name: %s with version: %s", name, serial)
	secret, err := secretStore.Get(name, serial)
	if err != nil {
		return err
	}

	jsonBytes, err := json.Marshal(secret)
	if err != nil {
		return fmt.Errorf("Error encoding secret: %v", err)
	}

	// Print secret to stdout
	fmt.Println(string(jsonBytes))
	return nil
}
