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

package cmd

import (
	"flag"
	"testing"

	"github.com/urfave/cli"
)

func TestGetDefaultLogLevel(t *testing.T) {
	flagSet := flag.NewFlagSet("ecs-secrets", 0)
	context := cli.NewContext(nil, flagSet, nil)
	loglevel := getLogLevel(context)
	if loglevel != loglevelInfo {
		t.Errorf("Inorrect loglevel set: %s", loglevel)
	}
}

func TestGetDebugLogLevel(t *testing.T) {
	flagSet := flag.NewFlagSet("ecs-secrets", 0)
	flagSet.Bool(debugFlag, true, "")
	context := cli.NewContext(nil, flagSet, nil)
	loglevel := getLogLevel(context)
	if loglevel != loglevelDebug {
		t.Errorf("Inorrect loglevel set: %s", loglevel)
	}
}

func TestGetLoglLevelReturnsInfoWhenDebugFlagIsSetToFalse(t *testing.T) {
	flagSet := flag.NewFlagSet("ecs-secrets", 0)
	flagSet.Bool(debugFlag, false, "")
	context := cli.NewContext(nil, flagSet, nil)
	loglevel := getLogLevel(context)
	if loglevel != loglevelInfo {
		t.Errorf("Inorrect loglevel set: %s", loglevel)
	}
}
