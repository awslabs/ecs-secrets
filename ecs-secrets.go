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

package main

import (
	"os"

	"github.com/awslabs/ecs-secrets/modules/cmd"
	"github.com/awslabs/ecs-secrets/modules/version"
	log "github.com/cihub/seelog"
	"github.com/urfave/cli"
)

func main() {
	defer log.Flush()

	app := cli.NewApp()
	app.Name = version.AppName
	app.Usage = "Command line interface for ECS Secrets"
	app.Version = version.Version
	app.Author = "Amazon Web Services"

	app.Commands = []cli.Command{
		cmd.SetupCommand(),
		cmd.CreateCommand(),
		cmd.FetchCommand(),
		cmd.RevokeCommand(),
		cmd.DaemonCommand(),
	}

	app.Run(os.Args)
}
