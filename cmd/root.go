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
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	apiv1 "k8s.io/api/core/v1"
)

var config *viper.Viper

var cfgFile string

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
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	config = viper.New()
	if cfgFile != "" { // enable ability to specify config file via flag
		config.SetConfigFile(cfgFile)
	}
	config.SetConfigType("yaml")
	config.SetEnvPrefix("pitayabot")
	config.AddConfigPath(".")
	config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	config.AutomaticEnv()

	// If a config file is found, read it in.
	if err := config.ReadInConfig(); err != nil {
		fmt.Printf("Config file %s failed to load: %s.\n", cfgFile, err.Error())
		panic("Failed to load config file")
	}
	fillDefaultValues(config)
}

func fillDefaultValues(config *viper.Viper) {
	defaultsMap := map[string]interface{}{
		"game":                 "",
		"prometheus.port":      9191,
		"server.host":          "localhost",
		"server.tls":           false,
		"storage.type":         "memory",
		"kubernetes.namespace": apiv1.NamespaceDefault,
	}

	for param := range defaultsMap {
		config.SetDefault(param, defaultsMap[param])
	}
}
