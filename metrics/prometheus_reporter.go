package metrics

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/topfreegames/pitaya/constants"
)

var (
	prometheusReporter *PrometheusReporter
	once               sync.Once
)

// PrometheusReporter reports metrics to prometheus
type PrometheusReporter struct {
	game                 string
	countReportersMap    map[string]*prometheus.CounterVec
	summaryReportersMap  map[string]*prometheus.SummaryVec
	histogramRepotersMap map[string]*prometheus.HistogramVec
	gaugeReportersMap    map[string]*prometheus.GaugeVec
}

func (p *PrometheusReporter) registerMetrics(constLabels map[string]string) {
	constLabels["game"] = p.game
	constLabels["clientType"] = "pitaya-bot"

	// HandlerResponseTimeMs summary
	p.summaryReportersMap[ResponseTime] = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:   fmt.Sprintf("pitaya_bot_%s", p.game),
			Subsystem:   "handler",
			Name:        ResponseTime,
			Help:        "the time to process a msg in nanoseconds",
			Objectives:  map[float64]float64{0.7: 0.02, 0.95: 0.005, 0.99: 0.001},
			ConstLabels: constLabels,
		},
		[]string{"route"},
	)

	p.histogramRepotersMap[ResponseTimeHistogram] = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace:   fmt.Sprintf("pitaya_bot_%s", p.game),
			Subsystem:   "handler",
			Name:        ResponseTimeHistogram,
			Help:        "histogram of the time to process a msg in nanoseconds",
			ConstLabels: constLabels,
		},
		[]string{"route"},
	)

	p.countReportersMap[ErrorCount] = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   fmt.Sprintf("pitaya_bot_%s", p.game),
			Subsystem:   "handler",
			Name:        ErrorCount,
			Help:        "the time to process a msg in nanoseconds",
			ConstLabels: constLabels,
		},
		[]string{"route"},
	)

	toRegister := make([]prometheus.Collector, 0)
	for _, c := range p.countReportersMap {
		toRegister = append(toRegister, c)
	}

	for _, c := range p.gaugeReportersMap {
		toRegister = append(toRegister, c)
	}

	for _, c := range p.summaryReportersMap {
		toRegister = append(toRegister, c)
	}

	prometheus.MustRegister(toRegister...)
}

func metricsReporterHandler(prometheusHandler http.Handler, postHandlerAction func()) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		prometheusHandler.ServeHTTP(w, r)
		postHandlerAction()
	})
}

// GetPrometheusReporter gets the prometheus reporter singleton
func GetPrometheusReporter(game string, port int, constLabels map[string]string, postMetricsScrapeAction func()) *PrometheusReporter {
	once.Do(func() {
		prometheusReporter = &PrometheusReporter{
			game:                 game,
			countReportersMap:    make(map[string]*prometheus.CounterVec),
			summaryReportersMap:  make(map[string]*prometheus.SummaryVec),
			histogramRepotersMap: make(map[string]*prometheus.HistogramVec),
			gaugeReportersMap:    make(map[string]*prometheus.GaugeVec),
		}

		prometheusReporter.registerMetrics(constLabels)
		http.Handle("/metrics", metricsReporterHandler(prometheus.Handler(), postMetricsScrapeAction))

		go (func() {
			log.Printf("Running prometheus on port %d for game %s", port, game)
			log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
		})()
	})
	return prometheusReporter
}

// ReportSummary reports a summary metric
//  - implements the ReportSummary method of the Reporter interface
func (p *PrometheusReporter) ReportSummary(metric string, labels map[string]string, value float64) error {
	sum := p.summaryReportersMap[metric]
	if sum != nil {
		sum.With(labels).Observe(value)
		return nil
	}
	return constants.ErrMetricNotKnown
}

// ReportHistogram reports a summary metric
//  - implements the ReportSummary method of the Reporter interface
func (p *PrometheusReporter) ReportHistogram(metric string, labels map[string]string, value float64) error {
	sum := p.histogramRepotersMap[metric]
	if sum != nil {
		sum.With(labels).Observe(value)
		return nil
	}
	return constants.ErrMetricNotKnown
}

// ReportCount reports a summary metric
//  - implements the ReportCount method of the Reporter interface
func (p *PrometheusReporter) ReportCount(metric string, labels map[string]string, count float64) error {
	cnt := p.countReportersMap[metric]
	if cnt != nil {
		cnt.With(labels).Add(count)
		return nil
	}
	return constants.ErrMetricNotKnown
}

// ReportGauge reports a gauge metric
//  - implements the ReportGauge method of the Reporter interface
func (p *PrometheusReporter) ReportGauge(metric string, labels map[string]string, value float64) error {
	g := p.gaugeReportersMap[metric]
	if g != nil {
		g.With(labels).Set(value)
		return nil
	}
	return constants.ErrMetricNotKnown
}
