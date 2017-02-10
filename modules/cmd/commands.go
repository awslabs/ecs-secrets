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

import "github.com/urfave/cli"

const (
	applicationNameFlag = "application-name"
	debugFlag           = "debug"

	createSecretsPrincipalFlag = "create-principal"
	fetchSecretsRoleFlag       = "fetch-role"
	nameFlag                   = "name"
	payloadFlag                = "payload"
	payloadLocationFlag        = "payload-location"
	serialFlag                 = "serial"
)

// appendCommonCLIFlags returns a modified list of flags by appending the
// common CLI flags to the list provided in function arguments. Not using
// global flags here because the cli package starts caring about the
// ordering/placement of global flags, which makes it awkward/non-intuitive
// to invoke commands.
//
// Example:
// ecs-secrets --debug create vs ecs-secrets create --debug
func appendCommonCLIFlags(flags []cli.Flag) []cli.Flag {
	return append(flags, []cli.Flag{
		cli.StringFlag{
			Name:  applicationNameFlag,
			Usage: "Specifies the name of the application.",
		},
		cli.BoolFlag{
			Name:  debugFlag,
			Usage: "Run in debug mode.",
		},
	}...)
}

func SetupCommand() cli.Command {
	return cli.Command{
		Name:   "setup",
		Usage:  "Sets up the resources needed for ECS Secrets.",
		Before: beforeCommand,
		Action: setupCommand,
		Flags: appendCommonCLIFlags([]cli.Flag{
			cli.StringFlag{
				Name:  createSecretsPrincipalFlag,
				Usage: "Specifies the IAM Principal for creating secrets.",
			},
			cli.StringFlag{
				Name:  fetchSecretsRoleFlag,
				Usage: "Specifies the IAM Role Arn for fetcing secrets.",
			},
		}),
	}
}

func CreateCommand() cli.Command {
	return cli.Command{
		Name:    "create",
		Aliases: []string{"store"},
		Usage:   "Creates a secret.",
		Before:  beforeCommand,
		Action:  createCommand,
		Flags: appendCommonCLIFlags([]cli.Flag{
			cli.StringFlag{
				Name:  nameFlag,
				Usage: "Specifies the name of the secret.",
			},
			cli.StringFlag{
				Name:  payloadFlag,
				Usage: "Specifies the payload for the secret.",
			},
			cli.StringFlag{
				Name:  payloadLocationFlag,
				Usage: "Specifies the file path containing the secrets payload.",
			},
		}),
	}
}

func FetchCommand() cli.Command {
	return cli.Command{
		Name:    "fetch",
		Aliases: []string{"get"},
		Usage:   "Gets a secret.",
		Before:  beforeCommand,
		Action:  fetchCommand,
		Flags: appendCommonCLIFlags([]cli.Flag{
			cli.StringFlag{
				Name:  nameFlag,
				Usage: "Specifies the name of the secret.",
			},
			cli.StringFlag{
				Name:  serialFlag,
				Usage: "Specifies the verison of the secret.",
			},
		}),
	}
}

func RevokeCommand() cli.Command {
	return cli.Command{
		Name:   "revoke",
		Usage:  "Revokes a secret.",
		Before: beforeCommand,
		Action: revokeCommand,
		Flags: appendCommonCLIFlags([]cli.Flag{
			cli.StringFlag{
				Name:  nameFlag,
				Usage: "Specifies the name of the secret.",
			},
			cli.StringFlag{
				Name:  serialFlag,
				Usage: "Specifies the verison of the secret.",
			},
		}),
	}
}

func DaemonCommand() cli.Command {
	return cli.Command{
		Name:   "daemon",
		Usage:  "Starts ECS Secrets daemon.",
		Before: beforeCommand,
		Action: daemonCommand,
		Flags:  appendCommonCLIFlags([]cli.Flag{}),
	}
}
