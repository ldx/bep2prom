package server

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	bes "google.golang.org/genproto/googleapis/devtools/build/v1"
	"google.golang.org/protobuf/types/known/emptypb"

	bb "github.com/ldx/bep2prom/pkg/build_event_stream"
	"github.com/ldx/bep2prom/pkg/metrics"
)

type Server struct {
	builds map[string]time.Time
}

func New() *Server {
	return &Server{
		builds: make(map[string]time.Time),
	}
}

// Handle lifecycle events.
func (s *Server) PublishLifecycleEvent(ctx context.Context, in *bes.PublishLifecycleEventRequest) (*emptypb.Empty, error) {
	ev := in.GetBuildEvent()
	if ev == nil {
		return &emptypb.Empty{}, nil
	}
	buildID := ev.GetStreamId().GetBuildId()
	if buildEnqueued := ev.GetEvent().GetBuildEnqueued(); buildEnqueued != nil {
		// Build enqueued, persist start time.
		s.builds[buildID] = ev.GetEvent().EventTime.AsTime()
	}
	if buildFinished := ev.GetEvent().GetBuildFinished(); buildFinished != nil {
		if startedAt, ok := s.builds[buildID]; ok {
			// Build finished, calculate duration.
			result := buildFinished.GetStatus().GetResult().String()
			duration := ev.GetEvent().EventTime.AsTime().Sub(startedAt)
			log.Printf("Build %s finished in %s: %s", buildID, duration, result)
			metrics.BuildCompleted.WithLabelValues(buildID, result).Observe(duration.Seconds())
			delete(s.builds, buildID)
		}
	}
	return &emptypb.Empty{}, nil
}

//PublishBuildToolEventStream(PublishBuildEvent_PublishBuildToolEventStreamServer) error
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
		if be, ok := in.OrderedBuildEvent.Event.Event.(*bes.BuildEvent_BazelEvent); ok {
			event := new(bb.BuildEvent)
			if err = be.BazelEvent.UnmarshalTo(event); err == nil {
				updateMetricsFromEvent(buildID, invocationID, sequenceNumber, event)
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

func updateMetricsFromEvent(buildID, invoicationID string, sequenceNumber int64, event *bb.BuildEvent) {
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
	desc := fmt.Sprintf("%s %s %d %s %s", buildID, invoicationID, sequenceNumber, kind, label)
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
		log.Printf("buildMetrics: %v\n", buildMetrics)
	}
	//if buildToolLogs := event.GetBuildToolLogs(); buildToolLogs != nil {
	//	log.Printf("buildToolLogs: %v\n", buildToolLogs)
	//}
	//if children := event.GetChildren(); children != nil {
	//	log.Printf("children: %v\n", children)
	//}
	if completed := event.GetCompleted(); completed != nil {
		log.Printf("completed: success %v\n", completed.GetSuccess())
		metrics.BuildEventCompleted.WithLabelValues(buildID, invoicationID, kind, label).Inc()
	}
	if configuration := event.GetConfiguration(); configuration != nil {
		log.Printf("configuration: %v\n", configuration)
		metrics.BuildEventConfiguration.WithLabelValues(buildID, invoicationID, configuration.Mnemonic, configuration.PlatformName, configuration.Cpu).Inc()
	}
	if configured := event.GetConfigured(); configured != nil {
		log.Printf("configured: %v\n", configured)
		metrics.BuildEventConfigured.WithLabelValues(buildID, invoicationID, configured.TargetKind).Inc()
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
		metrics.BuildEventFinished.WithLabelValues(buildID, invoicationID, overallSuccess).Inc()
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
		metrics.BuildEventStarted.WithLabelValues(buildID, invoicationID, started.BuildToolVersion, started.Command).Inc()
	}
	//if structuredCommandLine := event.GetStructuredCommandLine(); structuredCommandLine != nil {
	//	log.Printf("structuredCommandLine: %v\n", structuredCommandLine)
	//}
	if targetSummary := event.GetTargetSummary(); targetSummary != nil {
		log.Printf("targetSummary: %v\n", targetSummary)
	}
	if testResult := event.GetTestResult(); testResult != nil {
		log.Printf("%s testResult: %v\n", desc, testResult)
		status := testResult.Status.String()
		cachedLocally := fmt.Sprintf("%v", testResult.CachedLocally)
		metrics.BuildEventTestResult.WithLabelValues(buildID, invoicationID, status, cachedLocally).Inc()
	}
	if testSummary := event.GetTestSummary(); testSummary != nil {
		log.Printf("testSummary: %v\n", testSummary)
		overallStatus := testSummary.GetOverallStatus().String()
		attemptCount := fmt.Sprintf("%d", testSummary.GetAttemptCount())
		runCount := fmt.Sprintf("%d", testSummary.GetRunCount())
		shardCount := fmt.Sprintf("%d", testSummary.GetShardCount())
		totalNumCached := fmt.Sprintf("%d", testSummary.GetTotalNumCached())
		totalRunCount := fmt.Sprintf("%d", testSummary.GetTotalRunCount())
		metrics.BuildEventTestSummaryOverallStatus.WithLabelValues(buildID, invoicationID, overallStatus).Inc()
		metrics.BuildEventTestSummaryAttemptCount.WithLabelValues(buildID, invoicationID, attemptCount).Inc()
		metrics.BuildEventTestSummaryRunCount.WithLabelValues(buildID, invoicationID, runCount).Inc()
		metrics.BuildEventTestSummaryShardCount.WithLabelValues(buildID, invoicationID, shardCount).Inc()
		metrics.BuildEventTestSummaryTotalNumCached.WithLabelValues(buildID, invoicationID, totalNumCached).Inc()
		metrics.BuildEventTestSummaryTotalRunCount.WithLabelValues(buildID, invoicationID, totalRunCount).Inc()
	}
	//if unstructuredCommandLine := event.GetUnstructuredCommandLine(); unstructuredCommandLine != nil {
	//	log.Printf("unstructuredCommandLine: %v\n", unstructuredCommandLine)
	//}
	//if workspaceInfo := event.GetWorkspaceInfo(); workspaceInfo != nil {
	//	log.Printf("workspaceInfo: %v\n", workspaceInfo)
	//}
	//if workspaceStatus := event.GetWorkspaceStatus(); workspaceStatus != nil {
	//	log.Printf("workspaceStatus: %v\n", workspaceStatus)
	//}
}
