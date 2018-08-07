package metrics

// Reporter interface
type Reporter interface {
	ReportCount(metric string, tags map[string]string, count float64) error
	ReportSummary(metric string, tags map[string]string, value float64) error
	ReportHistogram(metric string, tags map[string]string, value float64) error
	ReportGauge(metric string, tags map[string]string, value float64) error
}
