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
	log "github.com/cihub/seelog"

	"github.com/awslabs/ecs-secrets/modules/store"
	"github.com/urfave/cli"
)

func revokeCommand(context *cli.Context) error {
	// Validate that application name has been specified
	appName, err := getRequiredArgumentFromFlag(context, applicationNameFlag)
	if err != nil {
		return err
	}
	return doRevoke(context, createSecretStore(appName))
}

func doRevoke(context *cli.Context, secretStore store.Store) error {
	name := context.String(nameFlag)
	// Validate that secrets name has been specified
	name, err := getRequiredArgumentFromFlag(context, nameFlag)
	if err != nil {
		return err
	}

	// Validate that secrets version has been specified
	serial, err := getRequiredArgumentFromFlag(context, serialFlag)
	if err != nil {
		return err
	}

	log.Debugf("Revoking secret name: %s with version:%s", name, serial)
	err = secretStore.Revoke(name, serial)
	if err != nil {
		return err
	}

	log.Infof("Revoked secret: %s", name)
	return nil
}
