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
	appsv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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

func getSpecs(specsDirectory string) ([]*models.Spec, error) {
	ret := make([]*models.Spec, 0)
	err := filepath.Walk(specsDirectory,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
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
			sleepDuration := time.Duration(random.Intn(1000)) * time.Millisecond
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
		err := runClients(app, spec, config, logger)
		if err != nil {
			compoundError = append(compoundError, err...)
		}

		elapsed := time.Now().UTC().Sub(start)
		if elapsed.Seconds() > duration {
			break
		}
	}

	return compoundError
}

// Launch launches the bot spec
func Launch(app *state.App, config *viper.Viper, specsDirectory string, duration float64, shouldReportMetrics bool) {
	log := logrus.New()
	log.Formatter = new(logrus.TextFormatter)
	log.Out = os.Stdout
	logger := log.WithFields(logrus.Fields{
		"source":   "pitaya-bot",
		"function": "launch",
	})

	specs, err := getSpecs(specsDirectory)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Infof("Found %d specs to be executed", len(specs))

	var wg sync.WaitGroup
	errmutex := sync.Mutex{}
	compoundError := []error{}
	for _, spec := range specs {
		wg.Add(1)
		go func(spec *models.Spec) {
			err := runSpec(app, spec, config, duration, logger)
			if err != nil {
				errmutex.Lock()
				compoundError = append(compoundError, err...)
				errmutex.Unlock()
			}
			wg.Done()
		}(spec)
	}

	wg.Wait()

	logger.Info("Finished running bots")
	app.FinishedExecition = true

	if shouldReportMetrics {
		logger.Info("Waiting for metrics to be collected...")
		select {
		case <-app.DieChan: // when dieChan is closed the application can quit
			<-time.After(10 * time.Second)
			logger.Info("DieChan closed - All done. Application will close")
		}
	}

	if len(compoundError) > 0 {
		logger.Error("Spec execution failed")
		logger.Error(compoundError)
		os.Exit(1)
	}
}

// LaunchKubernetes launches the manager to create the pods to run specs
func LaunchKubernetes(app *state.App, config *viper.Viper, specsDirectory string, duration float64, shouldReportMetrics bool) {
	log := logrus.New()
	log.Formatter = new(logrus.TextFormatter)
	log.Out = os.Stdout
	logger := log.WithFields(logrus.Fields{
		"source":   "pitaya-bot",
		"function": "launchKubernetes",
	})

	specs, err := getSpecs(specsDirectory)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Infof("Found %d specs to be executed", len(specs))

	clusterConfig, err := rest.InClusterConfig()
	if err != nil {
		logger.Fatal(err)
	}

	clientset, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Infof("Kubernetes In Cluster Client created")

	namespaces := make([]string, len(specs))
	for i := 0; i < len(specs); i++ {
		namespaces[i] = fmt.Sprintf("pitaya-bot-%v", i)
	}

	configMapClient := clientset.CoreV1().ConfigMaps(config.GetString("kubernetes.namespace"))
	deploymentsClient := clientset.BatchV1().Jobs(config.GetString("kubernetes.namespace"))

	configBinary, err := ioutil.ReadFile(config.ConfigFileUsed())
	if err != nil {
		logger.Fatal(err)
	}

	configMap := &apiv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "config",
			Labels: map[string]string{
				"app":  "pitaya-bot-pod",
				"game": config.GetString("game"),
			},
		},
		BinaryData: map[string][]byte{"config.yaml": configBinary},
	}

	if _, err = configMapClient.Create(configMap); err != nil {
		logger.Fatal(err)
	}
	logger.Infof("Created configMap config.yaml")

	for index, spec := range specs {
		specBinary, err := ioutil.ReadFile(spec.Name)
		if err != nil {
			logger.Fatal(err)
		}

		configMap = &apiv1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespaces[index],
				Labels: map[string]string{
					"app":  "pitaya-bot-pod",
					"game": config.GetString("game"),
				},
			},
			BinaryData: map[string][]byte{"spec.json": specBinary},
		}

		if _, err = configMapClient.Create(configMap); err != nil {
			logger.Fatal(err)
		}
		logger.Infof("Created config map %s", namespaces[index])

		deployment := &appsv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pitaya-bot-pod",
			},
			Spec: appsv1.JobSpec{
				//Parallelism: int32Ptr(1), TODO: Via config file, see how many bots are to be instantiated
				Completions: int32Ptr(1),
				Template: apiv1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app":  "pitaya-bot",
							"game": config.GetString("game"),
						},
					},
					Spec: apiv1.PodSpec{
						RestartPolicy: apiv1.RestartPolicyOnFailure,
						Containers: []apiv1.Container{
							{
								Name:    "pitaya-bot-pod",
								Image:   "pitaya-bot",
								Command: []string{"pitaya-bot"},
								Args:    []string{"run"},
							},
						},
						Volumes: []apiv1.Volume{
							{
								Name: "spec",
								VolumeSource: apiv1.VolumeSource{
									ConfigMap: &apiv1.ConfigMapVolumeSource{
										LocalObjectReference: apiv1.LocalObjectReference{Name: namespaces[index]},
									},
								},
							},
							{
								Name: "config",
								VolumeSource: apiv1.VolumeSource{
									ConfigMap: &apiv1.ConfigMapVolumeSource{
										LocalObjectReference: apiv1.LocalObjectReference{Name: "config"},
									},
								},
							},
						},
					},
				},
			},
		}

		if _, err := deploymentsClient.Create(deployment); err != nil {
			logger.Fatal(err)
		}
		logger.Infof("Created pod %s", namespaces[index])
	}

	logger.Info("Finished instantiating bots")
	app.FinishedExecition = true
}

func int32Ptr(i int32) *int32 { return &i }
