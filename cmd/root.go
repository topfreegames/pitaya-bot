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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/homedir"
)

var (
	config  *viper.Viper
	cfgFile string
	verbose int
	logJSON bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pitaya-bot",
	Short: "A Pitaya bot",
	Long:  `A Pitaya bot useful for testing paths and doing stress tests`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "./config/config.yaml", "config file")
	rootCmd.PersistentFlags().IntVarP(
		&verbose, "verbose", "v", 2,
		"Verbosity level => v0: Error, v1=Warning, v2=Info, v3=Debug",
	)
	rootCmd.PersistentFlags().BoolVarP(&logJSON, "logJSON", "j", false, "logJSON output mode")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	config = CreateConfig(cfgFile)
}

// CreateConfig returns the config file all set up and ready to use
func CreateConfig(path string) *viper.Viper {
	config := viper.New()
	if path != "" { // enable ability to specify config file via flag
		config.SetConfigFile(path)
	}
	config.SetConfigType("yaml")
	config.SetEnvPrefix("pitayabot")
	config.AddConfigPath(".")
	config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	config.AutomaticEnv()

	// If a config file is found, read it in.
	if err := config.ReadInConfig(); err != nil {
		fmt.Printf("Config file %s failed to load: %s.\n", path, err.Error())
		return nil
	}
	fillDefaultValues(config)
	return config
}

func fillDefaultValues(config *viper.Viper) {
	defaultsMap := map[string]interface{}{
		"game":                                "",
		"prometheus.port":                     9191,
		"server.host":                         "localhost",
		"server.tls":                          "false",
		"server.serializer":                   "json",
		"server.protobuffer.docs":             "connector.docsHandler.docs",
		"storage.type":                        "memory",
		"kubernetes.config":                   filepath.Join(homedir.HomeDir(), ".kube", "config"),
		"kubernetes.context":                  "",
		"kubernetes.cpu":                      "250m",
		"kubernetes.image":                    "tfgco/pitaya-bot:latest",
		"kubernetes.imagepull":                "Always",
		"kubernetes.masterurl":                "",
		"kubernetes.memory":                   "256Mi",
		"kubernetes.namespace":                corev1.NamespaceDefault,
		"kubernetes.job.retry":                0,
		"manager.maxrequeues":                 5,
		"manager.wait":                        "1s",
		"bot.operation.maxSleep":              "500ms",
		"bot.operation.stopOnError":           false,
		"bot.spec.parallelism":                1,
		"custom.redis.pre.url":                "redis://localhost:9010",
		"custom.redis.pre.connectionTimeout":  10,
		"custom.redis.pre.script":             "",
		"custom.redis.post.url":               "redis://localhost:9010",
		"custom.redis.post.connectionTimeout": 10,
		"custom.redis.post.script":            "",
	}

	for param := range defaultsMap {
		config.SetDefault(param, defaultsMap[param])
	}
}
