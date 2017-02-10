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

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/awslabs/ecs-secrets/modules/cache"
	"github.com/awslabs/ecs-secrets/modules/crypt"
	"github.com/awslabs/ecs-secrets/modules/dao"
	"github.com/awslabs/ecs-secrets/modules/logger"
	"github.com/awslabs/ecs-secrets/modules/store"
	"github.com/urfave/cli"
)

const (
	loglevelInfo  = "info"
	loglevelDebug = "debug"
)

func beforeCommand(context *cli.Context) error {
	logger.SetupLogger(getLogLevel(context))
	return nil
}

func getLogLevel(context *cli.Context) string {
	if context.Bool(debugFlag) {
		return loglevelDebug
	}

	return loglevelInfo
}

func getRequiredArgumentFromFlag(context *cli.Context, flagName string) (string, error) {
	argValue := context.String(flagName)
	if argValue == "" {
		return "", fmt.Errorf("Missing required argument '%s'", flagName)
	}

	return argValue, nil
}

func createSecretStore(appName string) store.Store {
	dao := dao.NewDAO(appName, dynamodb.New(session.New()))
	lruCache := cache.NewLRUCache(cache.KeyCacheSize, cache.KeyCacheTTL)
	crypter := crypt.NewCrypter(kms.New(session.New()), lruCache, appName)
	return store.NewStore(appName, dao, crypter)
}
