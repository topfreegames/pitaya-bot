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
	game := config.GetString("game")
	prometheusPort := config.GetInt("prometheus.port")

	app := &App{
		FinishedExecition: false,
		DieChan:           make(chan struct{}),
	}

	if shouldReportMetrics {
		fmt.Println("[INFO] Will report metrics")
		mr := []metrics.Reporter{
			metrics.GetPrometheusReporter(game, prometheusPort, map[string]string{}, func() {
				defer app.Mu.Unlock()
				app.Mu.Lock()
				if app.FinishedExecition && !app.ChannelClosed {
					app.ChannelClosed = true
					close(app.DieChan)
				}
			}),
		}
		app.MetricsReporter = mr
	}

	return app
}
