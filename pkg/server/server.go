package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	bes "google.golang.org/genproto/googleapis/devtools/build/v1"
	"google.golang.org/protobuf/types/known/emptypb"

	bb "github.com/ldx/bep2prom/pkg/build_event_stream"
	"github.com/ldx/bep2prom/pkg/metrics"
)

var (
	stripMetadataPrefixes = map[string]bool{
		"STABLE_": true,
	}
)

type BuildInfo struct {
	invocationID string
	startedAt    time.Time         // Build start time.
	metadata     map[string]string // Workspace status variables.
}

type Server struct {
	builds map[string]*BuildInfo // Build ID + invocation ID -> build info.
}

func New() *Server {
	return &Server{
		builds: make(map[string]*BuildInfo),
	}
}

// PublishLifecycleEvent handles lifecycle events.
func (s *Server) PublishLifecycleEvent(ctx context.Context, in *bes.PublishLifecycleEventRequest) (*emptypb.Empty, error) {
	ev := in.GetBuildEvent()
	if ev == nil {
		return &emptypb.Empty{}, nil
	}
	buildID := ev.GetStreamId().GetBuildId()
	if buildEnqueued := ev.GetEvent().GetBuildEnqueued(); buildEnqueued != nil {
		// Build enqueued, persist start time.
		startedAt := ev.GetEvent().EventTime.AsTime()
		s.builds[buildID] = &BuildInfo{
			startedAt: startedAt,
			metadata:  make(map[string]string),
		}
		log.Printf("Build enqueued %v %v", buildID, startedAt)
	}
	if buildFinished := ev.GetEvent().GetBuildFinished(); buildFinished != nil {
		if buildInfo, ok := s.builds[buildID]; ok {
			// Build finished, calculate duration.
			result := buildFinished.GetStatus().GetResult().String()
			duration := ev.GetEvent().EventTime.AsTime().Sub(buildInfo.startedAt)
			log.Printf("Build %v finished in %s: %s", buildID, duration, result)
			labels := metrics.BuildLabels(buildID, "", buildInfo.metadata)
			metrics.BuildCompleted.With(metrics.MergeLabels(labels, map[string]string{"result": result})).Observe(duration.Seconds())
			delete(s.builds, buildID)
		}
	}
	return &emptypb.Empty{}, nil
}

// PublishBuildToolEventStream processes a stream of BEP events.
func (s *Server) PublishBuildToolEventStream(stream bes.PublishBuildEvent_PublishBuildToolEventStreamServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Printf("PublishBuildToolEventStream error: %v\n", err)
			return err
		}
		streamID := in.GetOrderedBuildEvent().GetStreamId()
		if streamID == nil {
			log.Printf("PublishBuildToolEventStream: no stream ID\n")
			continue
		}
		buildID := streamID.GetBuildId()
		invocationID := streamID.GetInvocationId()
		sequenceNumber := in.GetOrderedBuildEvent().GetSequenceNumber()
		var metadata map[string]string
		buildInfo := s.builds[buildID]
		if buildInfo != nil {
			metadata = buildInfo.metadata
		}
		if be, ok := in.OrderedBuildEvent.Event.Event.(*bes.BuildEvent_BazelEvent); ok {
			event := new(bb.BuildEvent)
			if err = be.BazelEvent.UnmarshalTo(event); err == nil {
				updateMetricsFromEvent(buildID, invocationID, sequenceNumber, metadata, event)
			} else {
				log.Printf("Warning, got BazelEvent that is not a BuildEvent: %s\n", be.BazelEvent.GetTypeUrl())
			}
		}
		resp := &bes.PublishBuildToolEventStreamResponse{
			StreamId:       streamID,
			SequenceNumber: sequenceNumber,
		}
		if err := stream.Send(resp); err != nil {
			log.Printf("Warning, sending response %v/%d: %v", resp.StreamId, resp.SequenceNumber, err)
		}
	}
}

func updateMetadataFromWorkspaceStatus(metadata map[string]string, workspaceStatus *bb.WorkspaceStatus) {
	for _, item := range workspaceStatus.Item {
		key := item.Key
		for prefix := range stripMetadataPrefixes {
			if len(key) > len(prefix) && key[:len(prefix)] == prefix {
				key = strings.ToLower(key[len(prefix):])
				break
			}
		}
		metadata[key] = item.Value
		log.Printf("Metadata: %s=%s", key, item.Value)
	}
}

func updateMetricsFromEvent(buildID, invocationID string, sequenceNumber int64, metadata map[string]string, event *bb.BuildEvent) {
	label := ""
	kind := ""
	if id := event.GetId(); id != nil {
		if targetCompleted := id.GetTargetCompleted(); targetCompleted != nil {
			label = targetCompleted.GetLabel()
			kind = "target"
		}
		if actionCompleted := id.GetActionCompleted(); actionCompleted != nil {
			label = actionCompleted.GetLabel()
			kind = "action"
		}
		if testResult := id.GetTestResult(); testResult != nil {
			label = testResult.GetLabel()
			kind = "test"
		}
	}
	// Metric labels.
	labels := metrics.BuildLabels(buildID, invocationID, metadata)
	// Description used for logging.
	desc := fmt.Sprintf("%v %d %s %s", labels, sequenceNumber, kind, label)
	log.Printf("=== BuildEvent: %s\n", desc)
	if action := event.GetAction(); action != nil {
		log.Printf("%s action: %v\n", desc, action)
	}
	if aborted := event.GetAborted(); aborted != nil {
		log.Printf("aborted: %v\n", aborted)
	}
	if buildMetadata := event.GetBuildMetadata(); buildMetadata != nil {
		log.Printf("buildMetadata: %v\n", buildMetadata)
	}
	if buildMetrics := event.GetBuildMetrics(); buildMetrics != nil {
		buf, err := json.Marshal(buildMetrics)
		if err != nil {
			log.Printf("buildMetrics: %v\n", buildMetrics)
		} else {
			log.Printf("buildMetrics: %s\n", buf)
		}
	}
	//if buildToolLogs := event.GetBuildToolLogs(); buildToolLogs != nil {
	//	log.Printf("buildToolLogs: %v\n", buildToolLogs)
	//}
	//if children := event.GetChildren(); children != nil {
	//	log.Printf("children: %v\n", children)
	//}
	if completed := event.GetCompleted(); completed != nil {
		log.Printf("completed: success %v\n", completed.GetSuccess())
		allLabels := metrics.MergeLabels(labels, map[string]string{"kind": kind, "label": label})
		metrics.BuildEventCompleted.With(allLabels).Inc()
	}
	if configuration := event.GetConfiguration(); configuration != nil {
		log.Printf("configuration: %v\n", configuration)
		allLabels := metrics.MergeLabels(labels, map[string]string{"mnemonic": configuration.GetMnemonic(), "platform_name": configuration.GetPlatformName(), "cpu": configuration.Cpu})
		metrics.BuildEventConfiguration.With(allLabels).Inc()
	}
	if configured := event.GetConfigured(); configured != nil {
		log.Printf("configured: %v\n", configured)
		allLabels := metrics.MergeLabels(labels, map[string]string{"kind": configured.GetTargetKind()})
		metrics.BuildEventConfigured.With(allLabels).Inc()
	}
	//if convenienceSymlinksIdentified := event.GetConvenienceSymlinksIdentified(); convenienceSymlinksIdentified != nil {
	//	log.Printf("convenienceSymlinksIdentified: %v\n", convenienceSymlinksIdentified)
	//}
	//if expanded := event.GetExpanded(); expanded != nil {
	//	log.Printf("expanded: %v\n", expanded)
	//}
	if fetch := event.GetFetch(); fetch != nil {
		log.Printf("fetch: %v\n", fetch)
	}
	if finished := event.GetFinished(); finished != nil {
		log.Printf("finished: %v\n", finished)
		overallSuccess := fmt.Sprintf("%v", finished.GetOverallSuccess())
		allLabels := metrics.MergeLabels(labels, map[string]string{"overall_success": overallSuccess})
		metrics.BuildEventFinished.With(allLabels).Inc()
	}
	//if lastMessage := event.GetLastMessage(); lastMessage {
	//	log.Printf("lastMessage: %v\n", lastMessage)
	//}
	//if namedSetOfFiles := event.GetNamedSetOfFiles(); namedSetOfFiles != nil {
	//	log.Printf("namedSetOfFiles: %v\n", namedSetOfFiles)
	//}
	if optionsParsed := event.GetOptionsParsed(); optionsParsed != nil {
		log.Printf("optionsParsed: %v\n", optionsParsed)
	}
	//if payload := event.GetPayload(); payload != nil {
	//	log.Printf("payload: %v\n", payload)
	//}
	//if progress := event.GetProgress(); progress != nil {
	//	log.Printf("progress: %v\n", progress)
	//}
	if started := event.GetStarted(); started != nil {
		log.Printf("started: %v\n", started)
		allLabels := metrics.MergeLabels(labels, map[string]string{"build_tool_version": started.GetBuildToolVersion(), "command": started.GetCommand()})
		metrics.BuildEventStarted.With(allLabels).Inc()
	}
	//if structuredCommandLine := event.GetStructuredCommandLine(); structuredCommandLine != nil {
	//	log.Printf("structuredCommandLine: %v\n", structuredCommandLine)
	//}
	if targetSummary := event.GetTargetSummary(); targetSummary != nil {
		log.Printf("targetSummary: %v\n", targetSummary)
	}
	if testResult := event.GetTestResult(); testResult != nil {
		buf, err := json.Marshal(testResult)
		if err != nil {
			log.Printf("%s testResult: %v\n", desc, testResult)
		} else {
			log.Printf("%s testResult: %s\n", desc, buf)
		}
		status := testResult.Status.String()
		cachedLocally := fmt.Sprintf("%v", testResult.CachedLocally)
		cachedRemotely := fmt.Sprintf("%v", testResult.ExecutionInfo != nil && testResult.ExecutionInfo.CachedRemotely || false)
		//testResult.ExecutionInfo.ExitCode
		//testResult.ExecutionInfo.ResourceUsage[0].Name
		//testResult.ExecutionInfo.ResourceUsage[0].Value
		allLabels := metrics.MergeLabels(labels, map[string]string{"status": status, "cached_locally": cachedLocally, "cached_remotely": cachedRemotely})
		metrics.BuildEventTestResult.With(allLabels).Inc()
	}
	if testSummary := event.GetTestSummary(); testSummary != nil {
		log.Printf("testSummary: %v\n", testSummary)
		overallStatus := testSummary.GetOverallStatus().String()
		metrics.BuildEventTestSummaryOverallStatus.With(metrics.MergeLabels(labels, map[string]string{"overall_status": overallStatus})).Inc()
		// Gauges.
		metrics.BuildEventTestSummaryAttemptCount.With(labels).Set(float64(testSummary.GetAttemptCount()))
		metrics.BuildEventTestSummaryRunCount.With(labels).Set(float64(testSummary.GetRunCount()))
		metrics.BuildEventTestSummaryShardCount.With(labels).Set(float64(testSummary.GetShardCount()))
		metrics.BuildEventTestSummaryTotalNumCached.With(labels).Set(float64(testSummary.GetTotalNumCached()))
		metrics.BuildEventTestSummaryTotalRunCount.With(labels).Set(float64(testSummary.GetTotalRunCount()))
	}
	//if unstructuredCommandLine := event.GetUnstructuredCommandLine(); unstructuredCommandLine != nil {
	//	log.Printf("unstructuredCommandLine: %v\n", unstructuredCommandLine)
	//}
	if workspaceInfo := event.GetWorkspaceInfo(); workspaceInfo != nil {
		log.Printf("workspaceInfo: %v\n", workspaceInfo)
	}
	if workspaceStatus := event.GetWorkspaceStatus(); workspaceStatus != nil {
		log.Printf("workspaceStatus: %v\n", workspaceStatus)
		if metadata != nil {
			updateMetadataFromWorkspaceStatus(metadata, workspaceStatus)
		} else {
			log.Printf("metadata is nil")
		}
	}
}
