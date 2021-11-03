package launcher

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/topfreegames/pitaya-bot/models"
	"github.com/topfreegames/pitaya-bot/runner"
	"github.com/topfreegames/pitaya-bot/state"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func readSpec(specPath string) (*models.Spec, error) {
	raw, err := ioutil.ReadFile(specPath)
	if err != nil {
		return nil, err
	}

	var spec models.Spec
	err = json.Unmarshal(raw, &spec)
	return &spec, err
}

func validFile(info os.FileInfo) bool {
	if info.IsDir() {
		return false
	}
	if runtime.GOOS != "windows" && info.Name()[0:1] == "." {
		return false
	}
	if strings.Contains(info.Name(), ".json") {
		return true
	}
	return false
}

func newKubernetesClientset(config *viper.Viper, logger logrus.FieldLogger) *kubernetes.Clientset {
	kubeConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: config.GetString("kubernetes.config")},
		&clientcmd.ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{Server: config.GetString("kubernetes.masterurl")}, CurrentContext: config.GetString("kubernetes.context")}).ClientConfig()
	if err != nil {
		logger.Fatal(err)
	}
	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		logger.Fatal(err)
	}
	return clientset
}

// GetSpecs will walk through specsDirectory and transform all spec JSONs into Spec objects
func GetSpecs(specsDirectory string) ([]*models.Spec, error) {
	ret := make([]*models.Spec, 0)
	err := filepath.Walk(specsDirectory,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() && strings.HasPrefix(info.Name(), "..") {
				return filepath.SkipDir
			}

			if !validFile(info) {
				return nil
			}
			spec, err := readSpec(path)
			if err != nil {
				return err
			}

			spec.Name = path
			ret = append(ret, spec)
			fmt.Println(path, info.Size())
			return nil
		})
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func runClients(app *state.App, spec *models.Spec, config *viper.Viper, logger logrus.FieldLogger) []error {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	var (
		errmutex      sync.Mutex
		wg            sync.WaitGroup
		compoundError []error
	)

	for i := 0; i < spec.NumberOfInstances; i++ {
		wg.Add(1)
		go func(i int) {
			sleepDuration := time.Duration(random.Intn(int(config.GetDuration("bot.operation.maxSleep"))))
			time.Sleep(sleepDuration)
			if err := runner.Run(app, config, spec, i, logger); err != nil {
				errmutex.Lock()
				compoundError = append(compoundError, err)
				errmutex.Unlock()
			}
			wg.Done()
		}(i)
	}

	wg.Wait()
	return compoundError
}

func runSpec(app *state.App, spec *models.Spec, config *viper.Viper, duration float64, logger logrus.FieldLogger) []error {
	logger = logger.WithFields(logrus.Fields{
		"spec": spec.Name,
	})

	logger.Debugf("Launching %d bots\n", spec.NumberOfInstances)

	var compoundError []error
	start := time.Now().UTC()
	for {
		if err := runClients(app, spec, config, logger); err != nil {
			compoundError = append(compoundError, err...)
			if config.GetBool("bot.operation.stopOnError") {
				break
			}
		}

		elapsed := time.Now().UTC().Sub(start)
		if elapsed.Seconds() > duration {
			break
		}
	}

	return compoundError
}

// Launch launches the bot spec
func Launch(
	app *state.App,
	config *viper.Viper,
	specsDirectory string,
	duration float64,
	shouldReportMetrics bool,
	logger logrus.FieldLogger,
) {
	logger = logger.WithFields(logrus.Fields{
		"function": "Launch",
	})

	specs, err := GetSpecs(specsDirectory)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Infof("Found %d specs to be executed", len(specs))

	var wg sync.WaitGroup
	errmutex := sync.Mutex{}
	compoundErrorHist := make(map[string]int)
	for _, spec := range specs {
		wg.Add(1)
		go func(spec *models.Spec) {
			if errs := runSpec(app, spec, config, duration, logger); errs != nil {
				errmutex.Lock()
				for _, err := range errs {
					compoundErrorHist[err.Error()]++
				}
				errmutex.Unlock()
			}
			wg.Done()
		}(spec)
	}

	wg.Wait()

	logger.Info("Finished running bots")
	app.FinishedExecution = true

	if shouldReportMetrics {
		logger.Info("Waiting for metrics to be collected...")
		select {
		case <-app.DieChan: // when dieChan is closed the application can quit
			<-time.After(10 * time.Second)
			logger.Info("DieChan closed - All done. Application will close")
		}
	}

	if len(compoundErrorHist) > 0 {
		logger.WithFields(logrus.Fields{
			"errors": compoundErrorHist,
		}).Error("Spec execution failed")
		os.Exit(1)
	}
}
