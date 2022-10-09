package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	BuildCompleted = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "build_completed_seconds",
			Help:    "Build completed time in seconds",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 20, 50, 100, 200, 500, 1000, 2000, 5000, 10000},
		},
		[]string{"build_id", "result"},
	)
	BuildEventStarted = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "build_event_started_total",
			Help: "Build event started total",
		},
		[]string{"build_id", "invocation_id", "build_tool_version", "command"},
	)
	BuildEventFinished = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "build_event_finished_total",
			Help: "Build event finished total",
		},
		[]string{"build_id", "invocation_id", "overall_success"},
	)
	BuildEventCompleted = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "build_event_completed_total",
			Help: "Build event completed total",
		},
		[]string{"build_id", "invocation_id", "kind", "label"},
	)
	BuildEventConfigured = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "build_event_configured_total",
			Help: "Build event configured total",
		},
		[]string{"build_id", "invocation_id", "target_kind"},
	)
	BuildEventConfiguration = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "build_event_configuration_total",
			Help: "Build event configuration total",
		},
		[]string{"build_id", "invocation_id", "mnemonic", "platform_name", "cpu"},
	)
	BuildEventTestResult = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "build_event_test_result_total",
			Help: "Build event test result total",
		},
		[]string{"build_id", "invocation_id", "status", "cached_locally"},
	)
	BuildEventTestSummaryOverallStatus = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "build_event_test_summary_total",
			Help: "Build event test summary total",
		},
		[]string{"build_id", "invocation_id", "overall_status"},
	)
	BuildEventTestSummaryAttemptCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "build_event_test_summary_attempt_count_total",
			Help: "Build event test summary attempt count total",
		},
		[]string{"build_id", "invocation_id", "attempt_count"},
	)
	BuildEventTestSummaryRunCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "build_event_test_summary_run_count_total",
			Help: "Build event test summary run count total",
		},
		[]string{"build_id", "invocation_id", "run_count"},
	)
	BuildEventTestSummaryShardCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "build_event_test_summary_shard_count_total",
			Help: "Build event test summary shard count total",
		},
		[]string{"build_id", "invocation_id", "shard_count"},
	)
	BuildEventTestSummaryTotalNumCached = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "build_event_test_summary_total_num_cached_total",
			Help: "Build event test summary total num cached total",
		},
		[]string{"build_id", "invocation_id", "total_num_cached"},
	)
	BuildEventTestSummaryTotalRunCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "build_event_test_summary_total_run_count_total",
			Help: "Build event test summary total run count total",
		},
		[]string{"build_id", "invocation_id", "total_run_count"},
	)
)

func init() {
	prometheus.MustRegister(BuildCompleted)
	prometheus.MustRegister(BuildEventStarted)
	prometheus.MustRegister(BuildEventFinished)
	prometheus.MustRegister(BuildEventCompleted)
	prometheus.MustRegister(BuildEventConfigured)
	prometheus.MustRegister(BuildEventConfiguration)
	prometheus.MustRegister(BuildEventTestResult)
	prometheus.MustRegister(BuildEventTestSummaryOverallStatus)
	prometheus.MustRegister(BuildEventTestSummaryAttemptCount)
	prometheus.MustRegister(BuildEventTestSummaryRunCount)
	prometheus.MustRegister(BuildEventTestSummaryShardCount)
	prometheus.MustRegister(BuildEventTestSummaryTotalNumCached)
	prometheus.MustRegister(BuildEventTestSummaryTotalRunCount)
}

func Serve() error {
	http.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(":2112", nil)
}
