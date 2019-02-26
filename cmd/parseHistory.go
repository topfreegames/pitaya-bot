// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/topfreegames/pitaya-bot/models"
)

var (
	historyFile   string
	historyOutput string
)

func parseHistory(logger logrus.FieldLogger, historyPath, outputPath string) error {
	logger = logger.WithFields(logrus.Fields{
		"operation":   "parseHistory",
		"historyPath": historyPath,
		"outputPath":  outputPath,
	})
	logger.Info("Parsing history file")
	f, err := os.Open(historyPath)
	if err != nil {
		return err
	}
	defer f.Close()

	spec := models.NewSpec("parsed_spec")
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		op, err := parseOperation(line)
		if err != nil {
			logger.WithError(err).Warnf("Failed to parse line: %s", line)
			continue
		}
		spec.SequentialOperations = append(spec.SequentialOperations, op)
	}

	bts, _ := json.MarshalIndent(spec, "", "  ")
	out, _ := os.Create(outputPath)
	out.Write(bts)
	out.Close()

	logger.Info("Finished parsing history file")
	return nil
}

func parseArg(arg interface{}) (map[string]interface{}, error) {
	var err error
	val := map[string]interface{}{
		"value": arg,
	}
	switch t := arg.(type) {
	case string:
		val["type"] = "string"
	case int:
		val["type"] = "int"
	case float64:
		val["type"] = "int"
	case bool:
		val["type"] = "bool"
	case map[string]interface{}:
		val["type"] = "object"
		val["value"], err = parseRequest(t)
		if err != nil {
			return nil, err
		}
	case []interface{}:
		val["type"] = "array"
		val["value"], err = parseRequest(t)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown type: %T", t)
	}
	return val, nil
}

func parseRequest(rawArgs interface{}) (interface{}, error) {
	switch args := rawArgs.(type) {
	case map[string]interface{}:
		ret := map[string]interface{}{}
		for k, v := range args {
			val, err := parseArg(v)
			if err != nil {
				return nil, err
			}
			ret[k] = val
		}
		return ret, nil

	case []interface{}:
		ret := []interface{}{}
		for _, v := range args {
			val, err := parseArg(v)
			if err != nil {
				return nil, err
			}
			ret = append(ret, val)
		}
		return ret, nil
	}
	return nil, fmt.Errorf("Unknown basic type")
}

func parseOperation(line string) (*models.Operation, error) {
	args := strings.Split(line, " ")
	if len(args) < 1 {
		return nil, errors.New("Empty operation")
	}
	switch args[0] {
	case "request":
		if len(args) < 3 {
			return nil, errors.New("Insufficient args for request")
		}

		data := []byte(strings.Join(args[2:], ""))
		var rawRequest map[string]interface{}
		if err := json.Unmarshal(data, &rawRequest); err != nil {
			return nil, err
		}

		requestArgs, err := parseRequest(rawRequest)
		if err != nil {
			return nil, err
		}
		a, ok := requestArgs.(map[string]interface{})
		if !ok {
			return nil, errors.New("Invalid type for operation args")
		}
		op := &models.Operation{
			Type: "request",
			URI:  args[1],
			Args: a,
		}

		return op, nil
	}
	return nil, errors.New("Unknown operation")
}

// parseHistoryCmd represents the parseHistory command
var parseHistoryCmd = &cobra.Command{
	Use:   "parseHistory",
	Short: "Parses a pitaya-cli history file to get commands",
	Long:  `Parses a pitaya-cli history file to get commands`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := getLogger()
		err := parseHistory(logger, historyFile, historyOutput)
		if err != nil {
			logger.WithError(err).Fatal("Failed to parse history file")
		}
	},
}

func init() {
	rootCmd.AddCommand(parseHistoryCmd)

	home, _ := homedir.Dir()
	historyPath := fmt.Sprintf("%s/.pitayacli_history", home)
	outputPath := fmt.Sprintf("spec_%d.json", time.Now().Unix())

	parseHistoryCmd.PersistentFlags().StringVarP(&historyFile, "history-path", "p", historyPath, "Path to history file")
	parseHistoryCmd.PersistentFlags().StringVarP(&historyOutput, "spec-output", "s", outputPath, "Path to save spec")
}
