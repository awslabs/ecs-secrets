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

package logger

import log "github.com/cihub/seelog"

func SetupLogger(loglevel string) {
	seelogLevel, ok := log.LogLevelFromString(loglevel)
	if !ok {
		seelogLevel = log.InfoLvl
	}
	consoleWriter, err := log.NewConsoleWriter()
	if err != nil {
		log.Error(err)
		return
	}
	formatter, err := log.NewFormatter("%UTCDate(2006-01-02T15:04:05Z07:00) [%LEVEL] %Msg%n")
	if err != nil {
		log.Error(err)
		return
	}
	root, err := log.NewSplitDispatcher(formatter, []interface{}{consoleWriter})
	if err != nil {
		log.Error(err)
		return
	}
	constraints, err := log.NewMinMaxConstraints(seelogLevel, log.CriticalLvl)
	if err != nil {
		log.Error(err)
		return
	}
	logger := log.NewAsyncLoopLogger(log.NewLoggerConfig(constraints, nil, root))
	log.ReplaceLogger(logger)
}
