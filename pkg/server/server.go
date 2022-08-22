package server

import (
	"context"
	"io"
	"log"

	bes "google.golang.org/genproto/googleapis/devtools/build/v1"
	"google.golang.org/protobuf/types/known/emptypb"

	bb "github.com/ldx/bep2prom/pkg/build_event_stream"
)

type Server struct {
}

//PublishLifecycleEvent(context.Context, *PublishLifecycleEventRequest) (*emptypb.Empty, error)
func (s *Server) PublishLifecycleEvent(ctx context.Context, in *bes.PublishLifecycleEventRequest) (*emptypb.Empty, error) {
	log.Printf("Lifecycle event: %v\n", in)
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
			return err
		}
		//log.Printf("Got via event stream: %v\n", in)
		if be, ok := in.OrderedBuildEvent.Event.Event.(*bes.BuildEvent_BazelEvent); ok {
			event := new(bb.BuildEvent)
			if err = be.BazelEvent.UnmarshalTo(event); err == nil {
				printDiagnostics(event)
			} else {
				log.Printf("Warning, got BazelEvent that is not a BuildEvent: %s\n", be.BazelEvent.GetTypeUrl())
			}
		}
		resp := &bes.PublishBuildToolEventStreamResponse{
			StreamId:       in.GetOrderedBuildEvent().GetStreamId(),
			SequenceNumber: in.GetOrderedBuildEvent().GetSequenceNumber(),
		}
		if err := stream.Send(resp); err != nil {
			log.Printf("Warning, sending response %v/%d: %v", resp.StreamId, resp.SequenceNumber, err)
		}
	}
}

func printDiagnostics(event *bb.BuildEvent) {
	if targetSummary := event.GetTargetSummary(); targetSummary != nil {
		log.Printf("Target summary: %v %s\n", targetSummary.OverallBuildSuccess, targetSummary.OverallTestStatus.String())
	}
	if testSummary := event.GetTestSummary(); testSummary != nil {
		log.Printf("Test summary: OverallStatus %s TotalRunCount %d RunCount %d AttemptCount %d ShardCount %d Passed %v Failed %v TotalNumCached %d FirstStartTime %d LastStopTime %d\n",
			testSummary.OverallStatus.String(),
			testSummary.TotalRunCount,
			testSummary.RunCount,
			testSummary.AttemptCount,
			testSummary.ShardCount,
			testSummary.Passed,
			testSummary.Failed,
			testSummary.TotalNumCached,
			testSummary.FirstStartTime,
			testSummary.LastStopTime)
	}
	if testResult := event.GetTestResult(); testResult != nil {
		log.Printf("Test result: Status %s StatusDetails %s CachedLocally %v TestAttemptStart %d TestAttemptDuration %d TestActionOutput %v Warning %v ExecutionInfo %v\n",
			testResult.Status.String(),
			testResult.StatusDetails,
			testResult.CachedLocally,
			testResult.TestAttemptStart.GetSeconds(),
			testResult.TestAttemptDuration.GetSeconds(),
			testResult.TestActionOutput,
			testResult.Warning,
			testResult.ExecutionInfo)
	}
}
