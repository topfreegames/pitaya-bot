// Copyright © 2018 TFG Co <backend@tfgco.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/topfreegames/pitaya-bot/launcher"
	"github.com/topfreegames/pitaya-bot/state"
)

var (
	specsDirectory  string
	pitayaBotType   string
	testDuration    time.Duration
	reportMetrics   bool
	deleteBeforeRun bool
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs the pitaya bot",
	Long:  `Runs the pitaya bot.`,
	Run: func(cmd *cobra.Command, args []string) {
		if config == nil {
			cmd.Help()
			os.Exit(0)
		}

		logger := getLogger()
		switch pitayaBotType {
		case "deploy-manager":
			launcher.LaunchManagerDeploy(config, specsDirectory, testDuration, reportMetrics, logger)
		case "remote-manager":
			launcher.LaunchRemoteManager(config, specsDirectory, testDuration, reportMetrics, deleteBeforeRun, logger)
		case "local-manager":
			launcher.LaunchLocalManager(config, specsDirectory, testDuration, reportMetrics, deleteBeforeRun, logger)
		case "delete-all":
			launcher.LaunchDeleteAll(config, logger)
		default:
			app := state.NewApp(config, reportMetrics)
			launcher.Launch(app, config, specsDirectory, testDuration.Seconds(), reportMetrics, logger)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// The bot will recusrively find all json files within specDir directory
	// and run the specs
	runCmd.PersistentFlags().StringVarP(&specsDirectory, "dir", "d", "./specs/", "Spec to run")
	runCmd.PersistentFlags().DurationVar(&testDuration, "duration", 1*time.Minute, "how long should the test take")
	runCmd.PersistentFlags().BoolVar(&reportMetrics, "report-metrics", false, "Should metrics be reported")
	runCmd.PersistentFlags().StringVarP(&pitayaBotType, "pitaya-bot-type", "t", "local", "Pitaya-Bot Type which will be run")
	runCmd.PersistentFlags().BoolVar(&deleteBeforeRun, "delete", false, "Delete all before run")
}
