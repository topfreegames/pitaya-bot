package state

import (
	"fmt"
	"sync"

	"github.com/spf13/viper"
	"github.com/topfreegames/pitaya-bot/metrics"
)

// App is the struct that holds the app global data shared between packages
type App struct {
	FinishedExecition bool
	ChannelClosed     bool
	DieChan           chan struct{}
	MetricsReporter   []metrics.Reporter
	Mu                sync.Mutex
}

// NewApp is the NewApp constructor
func NewApp(config *viper.Viper, shouldReportMetrics bool) *App {
	app := &App{
		FinishedExecition: false,
		DieChan:           make(chan struct{}),
	}

	if shouldReportMetrics {
		fmt.Println("[INFO] Will report metrics")
		app.MetricsReporter = []metrics.Reporter{
			metrics.GetPrometheusReporter(config.GetString("game"),
				config.GetInt("prometheus.port"),
				map[string]string{},
				func() {
					defer app.Mu.Unlock()
					app.Mu.Lock()
					if app.FinishedExecition && !app.ChannelClosed {
						app.ChannelClosed = true
						close(app.DieChan)
					}
				},
			)}
	}

	return app
}
