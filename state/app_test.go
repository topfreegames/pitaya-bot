package state

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestNewApp(t *testing.T) {
	var assertTypeTable = map[string]struct {
		config              *viper.Viper
		shouldReportMetrics bool
	}{
		"without_report": {viper.New(), false},
		"with_report":    {viper.New(), true},
	}

	for name, table := range assertTypeTable {
		t.Run(name, func(t *testing.T) {
			app := NewApp(table.config, table.shouldReportMetrics)
			assert.Equal(t, false, app.ChannelClosed)
			assert.Equal(t, false, app.FinishedExecution)
			assert.Empty(t, app.DieChan)
			if table.shouldReportMetrics {
				assert.NotEmpty(t, app.MetricsReporter)
			} else {
				assert.Empty(t, app.MetricsReporter)
			}
		})
	}
}
