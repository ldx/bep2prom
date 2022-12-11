package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	AllowedMetadataLabels = map[string]bool{
		"git_branch": true,
	}
	BuildCompleted = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "build_completed_seconds",
			Help:    "Build completed time in seconds",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 20, 50, 100, 200, 500, 1000, 2000, 5000, 10000},
		},
		[]string{"build_id", "invocation_id", "git_branch", "result"},
	)
	BuildEventStarted = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "build_event_started_total",
			Help: "Build event started total",
		},
		[]string{"build_id", "invocation_id", "git_branch", "build_tool_version", "command"},
	)
	BuildEventFinished = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "build_event_finished_total",
			Help: "Build event finished total",
		},
		[]string{"build_id", "invocation_id", "git_branch", "overall_success"},
	)
	BuildEventCompleted = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "build_event_completed_total",
			Help: "Build event completed total",
		},
		[]string{"build_id", "invocation_id", "git_branch", "kind", "label"},
	)
	BuildEventConfigured = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "build_event_configured_total",
			Help: "Build event configured total",
		},
		[]string{"build_id", "invocation_id", "git_branch", "target_kind"},
	)
	BuildEventConfiguration = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "build_event_configuration_total",
			Help: "Build event configuration total",
		},
		[]string{"build_id", "invocation_id", "git_branch", "mnemonic", "platform_name", "cpu"},
	)
	BuildEventTestResult = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "build_event_test_result_total",
			Help: "Build event test result total",
		},
		[]string{"build_id", "invocation_id", "git_branch", "status", "cached_locally"},
	)
	BuildEventTestSummaryOverallStatus = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "build_event_test_summary_total",
			Help: "Build event test summary total",
		},
		[]string{"build_id", "invocation_id", "git_branch", "overall_status"},
	)
	BuildEventTestSummaryAttemptCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "build_event_test_summary_attempt_count",
			Help: "Build event test summary attempt count",
		},
		[]string{"build_id", "invocation_id", "git_branch"},
	)
	BuildEventTestSummaryRunCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "build_event_test_summary_run_count",
			Help: "Build event test summary run count",
		},
		[]string{"build_id", "invocation_id", "git_branch"},
	)
	BuildEventTestSummaryShardCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "build_event_test_summary_shard_count",
			Help: "Build event test summary shard count",
		},
		[]string{"build_id", "invocation_id", "git_branch"},
	)
	BuildEventTestSummaryTotalNumCached = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "build_event_test_summary_total_num_cached",
			Help: "Build event test summary total num cached",
		},
		[]string{"build_id", "invocation_id", "git_branch"},
	)
	BuildEventTestSummaryTotalRunCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "build_event_test_summary_total_run_count",
			Help: "Build event test summary total run count",
		},
		[]string{"build_id", "invocation_id", "git_branch"},
	)
)

func BuildLabels(buildID, invocationID string, metadata map[string]string) prometheus.Labels {
	labels := prometheus.Labels{
		"build_id":      buildID,
		"invocation_id": invocationID,
		"git_branch":    "",
	}
	for k := range AllowedMetadataLabels {
		labels[k] = metadata[k]
	}
	return labels
}

func MergeLabels(labels ...prometheus.Labels) prometheus.Labels {
	merged := prometheus.Labels{}
	for _, l := range labels {
		for k, v := range l {
			merged[k] = v
		}
	}
	return merged
}

func register() {
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
	register()
	http.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(":2112", nil)
}
