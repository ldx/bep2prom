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
		[]string{"git_branch", "result"},
	)
	BuildEventStarted = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "build_event_started_total",
			Help: "Build event started total",
		},
		[]string{"git_branch", "build_tool_version", "command"},
	)
	BuildEventFinished = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "build_event_finished_total",
			Help: "Build event finished total",
		},
		[]string{"git_branch", "overall_success"},
	)
	BuildEventCompleted = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "build_event_completed_total",
			Help: "Build event completed total",
		},
		[]string{"git_branch", "kind", "label"},
	)
	BuildEventConfigured = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "build_event_configured_total",
			Help: "Build event configured total",
		},
		[]string{"git_branch", "kind"},
	)
	BuildEventConfiguration = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "build_event_configuration_total",
			Help: "Build event configuration total",
		},
		[]string{"git_branch", "mnemonic", "platform_name", "cpu"},
	)
	BuildEventTestResult = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "build_event_test_result_total",
			Help: "Build event test result total",
		},
		[]string{"git_branch", "status", "cached_locally", "cached_remotely", "strategy"},
	)
	BuildEventTestResultDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "build_event_test_result_duration_seconds",
			Help:    "Build event test result duration in seconds",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 20, 50, 100, 200, 500, 1000, 2000, 5000},
		},
		[]string{"git_branch", "status", "cached_locally", "cached_remotely", "strategy"},
	)
	BuildEventTestSummaryOverallStatus = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "build_event_test_summary_total",
			Help: "Build event test summary total",
		},
		[]string{"git_branch", "overall_status"},
	)
	BuildEventTestSummaryAttemptCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "build_event_test_summary_attempt_count",
			Help: "Build event test summary attempt count",
		},
		[]string{"git_branch"},
	)
	BuildEventTestSummaryRunCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "build_event_test_summary_run_count",
			Help: "Build event test summary run count",
		},
		[]string{"git_branch"},
	)
	BuildEventTestSummaryShardCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "build_event_test_summary_shard_count",
			Help: "Build event test summary shard count",
		},
		[]string{"git_branch"},
	)
	BuildEventTestSummaryTotalNumCached = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "build_event_test_summary_total_num_cached",
			Help: "Build event test summary total num cached",
		},
		[]string{"git_branch"},
	)
	BuildEventTestSummaryTotalRunCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "build_event_test_summary_total_run_count",
			Help: "Build event test summary total run count",
		},
		[]string{"git_branch"},
	)
	NumAnalyses = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "build_num_analyses",
			Help: "Number of analyses; if <= 1 it is a clean build",
		},
		[]string{"git_branch"},
	)
	PackagesLoaded = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "build_packages_loaded",
			Help: "Number of packages loaded",
		},
		[]string{"git_branch"},
	)
	TargetsConfigured = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "build_targets_configured",
			Help: "Number of targets configured",
		},
		[]string{"git_branch"},
	)
	ActionsCreated = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "build_actions_created",
			Help: "Number of actions created",
		},
		[]string{"git_branch"},
	)
	ActionsExecuted = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "build_actions_executed",
			Help: "Number of actions executed",
		},
		[]string{"git_branch"},
	)
	ActionDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "build_action_duration_seconds",
			Help:    "Action duration in seconds",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 20, 50, 100, 200, 500, 1000, 2000, 5000},
		},
		[]string{"git_branch", "mnemonic"},
	)
	OutputArtifactCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "build_output_artifact_count",
			Help: "Number of output artifacts",
		},
		[]string{"git_branch"},
	)
)

func BuildLabels(metadata map[string]string) prometheus.Labels {
	labels := prometheus.Labels{
		"git_branch": "",
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
	prometheus.MustRegister(BuildEventTestResultDuration)
	prometheus.MustRegister(BuildEventTestSummaryOverallStatus)
	prometheus.MustRegister(BuildEventTestSummaryAttemptCount)
	prometheus.MustRegister(BuildEventTestSummaryRunCount)
	prometheus.MustRegister(BuildEventTestSummaryShardCount)
	prometheus.MustRegister(BuildEventTestSummaryTotalNumCached)
	prometheus.MustRegister(BuildEventTestSummaryTotalRunCount)
	prometheus.MustRegister(NumAnalyses)
	prometheus.MustRegister(PackagesLoaded)
	prometheus.MustRegister(TargetsConfigured)
	prometheus.MustRegister(ActionsCreated)
	prometheus.MustRegister(ActionsExecuted)
	prometheus.MustRegister(ActionDuration)
	prometheus.MustRegister(OutputArtifactCount)
}

func Serve() error {
	register()
	http.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(":2112", nil)
}
