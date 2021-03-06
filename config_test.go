//
// Copyright (c) 2018-2019 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package main

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	assert := assert.New(t)

	testLogLevel := logrus.DebugLevel

	expectedConfig := agentConfig{
		logLevel: testLogLevel,
	}

	config := newConfig(testLogLevel)

	assert.True(reflect.DeepEqual(config, expectedConfig),
		"Config structures should be identical: got %+v, expecting %+v",
		config, expectedConfig)
}

func TestParseCmdlineOptionEmptyOption(t *testing.T) {
	assert := assert.New(t)

	a := &agentConfig{}

	err := a.parseCmdlineOption("")
	assert.NoError(err, "%v", err)
}

func TestParseCmdlineOptionWrongOptionValue(t *testing.T) {
	assert := assert.New(t)

	a := &agentConfig{}

	wrongOption := logLevelFlag + "=debgu"

	err := a.parseCmdlineOption(wrongOption)
	assert.Errorf(err, "Parsing should fail because wrong option %q", wrongOption)
}

func TestParseCmdlineOptionWrongOptionParam(t *testing.T) {
	assert := assert.New(t)

	a := &agentConfig{}

	wrongOption := "agent.lgo=debug"

	err := a.parseCmdlineOption(wrongOption)
	assert.Errorf(err, "Parsing should fail because wrong option %q", wrongOption)
}

func TestParseCmdlineOptionCorrectOptions(t *testing.T) {
	assert := assert.New(t)

	a := &agentConfig{}

	logFlagList := []string{"debug", "info", "warn", "error", "fatal", "panic"}

	for _, logFlag := range logFlagList {
		debug = false
		option := logLevelFlag + "=" + logFlag

		err := a.parseCmdlineOption(option)
		assert.NoError(err, "%v", err)

		if logFlag == "debug" {
			assert.True(debug)
		}
	}
}

func TestParseCmdlineOptionIncorrectOptions(t *testing.T) {
	assert := assert.New(t)

	a := &agentConfig{}

	logFlagList := []string{"debg", "ifo", "wan", "eror", "ftal", "pnic"}

	for _, logFlag := range logFlagList {
		option := logLevelFlag + "=" + logFlag

		err := a.parseCmdlineOption(option)
		assert.Errorf(err, "Should fail because of incorrect option %q", logFlag)
	}
}

func TestParseCmdlineOptionDevMode(t *testing.T) {
	assert := assert.New(t)

	a := &agentConfig{}

	type testData struct {
		option               string
		expectDevModeEnabled bool
	}

	data := []testData{
		{"agent.Devmode", false},
		{"agent.DevMode", false},
		{"devmode", false},
		{"DevMode", false},
		{"agent.devmodel", false},
		{"agent.devmode.", false},
		{"agent.devmode-", false},
		{"agent.devmode:", false},

		{"agent.devmode", true},
	}

	for i, d := range data {
		debug = false
		crashOnError = false

		err := a.parseCmdlineOption(d.option)
		assert.NoError(err)

		if !d.expectDevModeEnabled {
			continue
		}

		assert.True(debug, "test %d (%+v)", i, d)
		assert.True(crashOnError, "test %d (%+v)", i, d)
	}
}

func TestGetConfigEmptyFileName(t *testing.T) {
	assert := assert.New(t)

	a := &agentConfig{}

	err := a.getConfig("")
	assert.Error(err, "Should fail because command line path is empty")
}

func TestGetConfigFilePathNotExist(t *testing.T) {
	assert := assert.New(t)

	a := &agentConfig{}

	tmpFile, err := ioutil.TempFile("", "test")
	assert.NoError(err, "%v", err)

	fileName := tmpFile.Name()
	tmpFile.Close()
	err = os.Remove(fileName)
	assert.NoError(err, "%v", err)

	err = a.getConfig(fileName)
	assert.Error(err, "Should fail because command line path does not exist")
}

func TestGetConfig(t *testing.T) {
	assert := assert.New(t)

	a := &agentConfig{}

	tmpFile, err := ioutil.TempFile("", "test")
	assert.NoError(err, "%v", err)
	fileName := tmpFile.Name()

	tmpFile.Write([]byte(logLevelFlag + "=info"))
	tmpFile.Close()

	defer os.Remove(fileName)

	err = a.getConfig(fileName)
	assert.NoError(err, "%v", err)

	assert.True(a.logLevel == logrus.InfoLevel,
		"Log levels should be identical: got %+v, expecting %+v",
		a.logLevel, logrus.InfoLevel)
}

func TestParseCmdlineOptionTracing(t *testing.T) {
	assert := assert.New(t)

	type testData struct {
		option              string
		expectTraceEnabled  bool
		expectCollatedTrace bool
	}

	data := []testData{
		{"", false, false},
		{"moo", false, false},
		{"." + traceModeFlag, false, false},
		{traceModeFlag + ".", false, false},
		{"x" + traceModeFlag, false, false},
		{traceModeFlag + "x", false, false},
		{"x" + traceModeFlag + "x", false, false},
		{"=" + traceModeFlag, false, false},
		{traceModeFlag + "=", false, false},

		{traceModeFlag, true, false},
		{traceModeFlag + "=" + traceValueIsolated, true, false},
		{traceModeFlag + "=" + traceValueCollated, true, true},

		{traceModeFlag + "=" + traceValueIsolated + "x", false, false},
		{traceModeFlag + "=" + traceValueCollated + "x", false, false},
	}

	for i, d := range data {
		// force reset
		tracing = false
		collatedTrace = false
		debug = false

		a := &agentConfig{}

		tmpFile, err := ioutil.TempFile("", "")
		assert.NoError(err)

		fileName := tmpFile.Name()
		defer os.Remove(fileName)

		tmpFile.Write([]byte(d.option))
		tmpFile.Close()

		assert.False(tracing)
		assert.False(collatedTrace)
		assert.False(debug)

		err = a.getConfig(fileName)
		assert.NoError(err)

		if d.expectTraceEnabled {
			assert.Truef(tracing, "test %d (%+v)", i, d)
		} else {
			assert.Falsef(tracing, "test %d (%+v)", i, d)
		}

		if d.expectCollatedTrace {
			assert.Truef(collatedTrace, "test %d (%+v)", i, d)
		} else {
			assert.Falsef(collatedTrace, "test %d (%+v)", i, d)
		}

		if d.expectTraceEnabled || d.expectCollatedTrace {
			assert.True(debug, "test %d (%+v)", i, d)
		}
	}
}

func TestEnableTracing(t *testing.T) {
	assert := assert.New(t)

	type testData struct {
		collatedTrace bool
	}

	data := []testData{
		{false},
		{true},
	}

	for i, d := range data {
		// force reset
		tracing = false
		collatedTrace = false
		debug = false

		enableTracing(d.collatedTrace)

		assert.True(debug, "test %d (%+v)", i, d)
		assert.True(tracing, "test %d (%+v)", i, d)

		if d.collatedTrace {
			assert.True(collatedTrace, "test %d (%+v)", i, d)
		} else {
			assert.False(collatedTrace, "test %d (%+v)", i, d)
		}
	}
}
