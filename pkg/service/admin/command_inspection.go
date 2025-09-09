package admin

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	adminv2 "github.com/digital-asset/dazl-client/v8/go/api/com/daml/ledger/api/v2/admin"
	"github.com/noders-team/go-daml/pkg/model"
)

type CommandInspection interface {
	GetCommandStatus(ctx context.Context, commandIDPrefix string, state model.CommandState, limit uint32) ([]*model.CommandStatus, error)
}

type commandInspection struct {
	client adminv2.CommandInspectionServiceClient
}

func NewCommandInspectionClient(conn *grpc.ClientConn) *commandInspection {
	client := adminv2.NewCommandInspectionServiceClient(conn)
	return &commandInspection{
		client: client,
	}
}

func (c *commandInspection) GetCommandStatus(ctx context.Context, commandIDPrefix string, state model.CommandState, limit uint32) ([]*model.CommandStatus, error) {
	req := &adminv2.GetCommandStatusRequest{
		CommandIdPrefix: commandIDPrefix,
		State:           commandStateToProto(state),
		Limit:           limit,
	}

	resp, err := c.client.GetCommandStatus(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get command status: %w", err)
	}

	return commandStatusFromProtos(resp.CommandStatus), nil
}

func commandStateToProto(state model.CommandState) adminv2.CommandState {
	switch state {
	case model.CommandStatePending:
		return adminv2.CommandState_COMMAND_STATE_PENDING
	case model.CommandStateSucceeded:
		return adminv2.CommandState_COMMAND_STATE_SUCCEEDED
	case model.CommandStateFailed:
		return adminv2.CommandState_COMMAND_STATE_FAILED
	default:
		return adminv2.CommandState_COMMAND_STATE_UNSPECIFIED
	}
}

func commandStateFromProto(state adminv2.CommandState) model.CommandState {
	switch state {
	case adminv2.CommandState_COMMAND_STATE_PENDING:
		return model.CommandStatePending
	case adminv2.CommandState_COMMAND_STATE_SUCCEEDED:
		return model.CommandStateSucceeded
	case adminv2.CommandState_COMMAND_STATE_FAILED:
		return model.CommandStateFailed
	default:
		return model.CommandStateUnspecified
	}
}

func commandStatusFromProto(pb *adminv2.CommandStatus) *model.CommandStatus {
	if pb == nil {
		return nil
	}

	cs := &model.CommandStatus{
		State: commandStateFromProto(pb.State),
	}

	if pb.Started != nil {
		t := pb.Started.AsTime()
		cs.Started = &t
	}

	if pb.Completed != nil {
		t := pb.Completed.AsTime()
		cs.Completed = &t
	}

	return cs
}

func commandStatusFromProtos(pbs []*adminv2.CommandStatus) []*model.CommandStatus {
	result := make([]*model.CommandStatus, len(pbs))
	for i, pb := range pbs {
		result[i] = commandStatusFromProto(pb)
	}
	return result
}
