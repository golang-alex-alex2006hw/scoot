package execution

import (
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	log "github.com/sirupsen/logrus"
	remoteexecution "github.com/twitter/scoot/bazel/remoteexecution"
	google_rpc_code "google.golang.org/genproto/googleapis/rpc/code"
	google_rpc_status "google.golang.org/genproto/googleapis/rpc/status"

	"github.com/twitter/scoot/bazel"
	"github.com/twitter/scoot/bazel/execution/bazelapi"
	bazelthrift "github.com/twitter/scoot/bazel/execution/bazelapi/gen-go/bazel"
	scootproto "github.com/twitter/scoot/common/proto"
	"github.com/twitter/scoot/sched"
	"github.com/twitter/scoot/scootapi/gen-go/scoot"
)

func marshalAny(pb proto.Message) (*any.Any, error) {
	pbAny, err := ptypes.MarshalAny(pb)
	if err != nil {
		s := fmt.Sprintf("Failed to marshal proto message %q as Any: %s", pb, err)
		log.Error(s)
		return nil, err
	}
	return pbAny, nil
}

func validateExecRequest(req *remoteexecution.ExecuteRequest) error {
	if req == nil {
		return fmt.Errorf("Unexpected nil execute request")
	}
	actionDigest := req.GetActionDigest()
	if !bazel.IsValidDigest(actionDigest.GetHash(), actionDigest.GetSizeBytes()) {
		return fmt.Errorf("Request action digest is invalid")
	}
	return nil
}

// Extract Scoot-related job fields from request to populate a JobDef, and pass through bazel request
func execReqToScoot(req *remoteexecution.ExecuteRequest) (
	result sched.JobDefinition, err error) {
	if err := validateExecRequest(req); err != nil {
		return result, err
	}

	// NOTE fixed to lowest priority in early stages of Bazel support
	// ExecuteRequests do not have priority values, but the Action portion
	// contains Platform Properties which can be used to specify arbitary server-side behavior.
	result.Priority = sched.P0
	result.Tasks = []sched.TaskDefinition{}

	// Populate TaskDef and Command. Note that Argv and EnvVars are set with placeholders for these requests,
	// per Bazel API this data must be made available by the client in the CAS before submitting this request.
	// To prevent increasing load and complexity in the Scheduler, this lookup is done at run time on the Worker
	// which is required to support CAS interactions.
	// ActionDigest is added for convenience and universal availability
	// ExecutionMetadata is seeded with current time of queueing
	now := time.Now()
	var task sched.TaskDefinition
	task.TaskID = fmt.Sprintf("%s_%s_%d", TaskIDPrefix, req.GetActionDigest(), now.Unix())
	task.Command.Argv = []string{CommandDefault}
	task.Command.EnvVars = make(map[string]string)
	task.Command.SnapshotID = bazel.SnapshotIDFromDigest(req.GetActionDigest())
	task.Command.ExecuteRequest = &bazelapi.ExecuteRequest{
		Request: req,
		ExecutionMetadata: &remoteexecution.ExecutedActionMetadata{
			QueuedTimestamp: scootproto.GetTimestampFromTime(now),
		},
	}

	result.Tasks = append(result.Tasks, task)
	return result, nil
}

func validateBzJobStatus(js *scoot.JobStatus) error {
	if len(js.GetTaskData()) > 1 || len(js.GetTaskStatus()) > 1 {
		return fmt.Errorf(
			"TaskData and/or TaskStatus of Bazel job status has len > 1. TaskData: %+v. TaskStatus: %+v",
			js.GetTaskData(), js.GetTaskStatus())
	}

	return nil
}

type runStatus struct {
	*scoot.RunStatus
}

// This is necessary b/c autogenerated thrift code doesn't nil check RunStatus before attempting
// to access its fields - leads to invalid memory address / nil pointer dereference panic
func (rs *runStatus) GetBazelResult() *bazelthrift.ActionResult_ {
	if rs == nil || rs.RunStatus == nil {
		return nil
	}
	return rs.GetBazelResult_()
}

// The Metadata Stage represents the high-level status about the task run
func runStatusToExecuteOperationMetadata_Stage(rs *runStatus) remoteexecution.ExecuteOperationMetadata_Stage {
	if rs == nil || rs.RunStatus == nil {
		return remoteexecution.ExecuteOperationMetadata_UNKNOWN
	}
	switch rs.Status {
	case scoot.RunStatusState_UNKNOWN:
		return remoteexecution.ExecuteOperationMetadata_UNKNOWN
	case scoot.RunStatusState_PENDING:
		return remoteexecution.ExecuteOperationMetadata_QUEUED
	case scoot.RunStatusState_RUNNING:
		return remoteexecution.ExecuteOperationMetadata_EXECUTING
	case scoot.RunStatusState_COMPLETE:
		return remoteexecution.ExecuteOperationMetadata_COMPLETED
	case scoot.RunStatusState_FAILED:
		return remoteexecution.ExecuteOperationMetadata_COMPLETED
	case scoot.RunStatusState_ABORTED:
		return remoteexecution.ExecuteOperationMetadata_COMPLETED
	case scoot.RunStatusState_TIMEDOUT:
		return remoteexecution.ExecuteOperationMetadata_COMPLETED
	case scoot.RunStatusState_BADREQUEST:
		return remoteexecution.ExecuteOperationMetadata_COMPLETED
	default:
		return remoteexecution.ExecuteOperationMetadata_UNKNOWN
	}
}

// The Google RPC Status is used in the ExecuteResponse and is used primarily by the client to
// determine the detailed state of a completed run. The resulting status codes are based on
// guidelines defined in the Bazel API unless otherwise noted.
func runStatusToGoogleRpcStatus(rs *runStatus) *google_rpc_status.Status {
	if rs == nil || rs.RunStatus == nil {
		return &google_rpc_status.Status{}
	}
	switch rs.Status {
	case scoot.RunStatusState_UNKNOWN:
		return &google_rpc_status.Status{
			Code: int32(google_rpc_code.Code_UNKNOWN),
		}
	case scoot.RunStatusState_PENDING:
		return &google_rpc_status.Status{
			Code: int32(google_rpc_code.Code_OK),
		}
	case scoot.RunStatusState_RUNNING:
		return &google_rpc_status.Status{
			Code: int32(google_rpc_code.Code_OK),
		}
	case scoot.RunStatusState_COMPLETE:
		return &google_rpc_status.Status{
			Code: int32(google_rpc_code.Code_OK),
		}
	case scoot.RunStatusState_FAILED:
		return &google_rpc_status.Status{
			Code: int32(google_rpc_code.Code_INTERNAL),
		}
	// NOTE: The API does not indicate that ABORTED as an acceptable error, however
	// given both the prevalence of Abort behavior in Scoot and the obviousness of the
	// given status code, it is appropriate that we deviate slightly here.
	case scoot.RunStatusState_ABORTED:
		return &google_rpc_status.Status{
			Code: int32(google_rpc_code.Code_ABORTED),
		}
	case scoot.RunStatusState_TIMEDOUT:
		return &google_rpc_status.Status{
			Code: int32(google_rpc_code.Code_DEADLINE_EXCEEDED),
		}
	case scoot.RunStatusState_BADREQUEST:
		return &google_rpc_status.Status{
			Code: int32(google_rpc_code.Code_INTERNAL),
		}
	default:
		return &google_rpc_status.Status{
			Code: int32(google_rpc_code.Code_UNKNOWN),
		}
	}
}

func runStatusToDoneBool(rs *runStatus) bool {
	if rs == nil || rs.RunStatus == nil {
		return false
	}
	switch rs.Status {
	case scoot.RunStatusState_UNKNOWN:
		return false
	case scoot.RunStatusState_PENDING:
		return false
	case scoot.RunStatusState_RUNNING:
		return false
	case scoot.RunStatusState_COMPLETE:
		return true
	case scoot.RunStatusState_FAILED:
		return true
	case scoot.RunStatusState_ABORTED:
		return true
	case scoot.RunStatusState_TIMEDOUT:
		return true
	case scoot.RunStatusState_BADREQUEST:
		return true
	default:
		return false
	}
}
