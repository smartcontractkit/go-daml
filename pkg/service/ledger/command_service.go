package ledger

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	v2 "github.com/digital-asset/dazl-client/v8/go/api/com/daml/ledger/api/v2"
	"github.com/noders-team/go-daml/pkg/model"
)

type CommandService interface {
	SubmitAndWait(ctx context.Context, req *model.SubmitAndWaitRequest) (*model.SubmitAndWaitResponse, error)
}

type commandService struct {
	client v2.CommandServiceClient
}

func NewCommandServiceClient(conn *grpc.ClientConn) *commandService {
	client := v2.NewCommandServiceClient(conn)
	return &commandService{
		client: client,
	}
}

func (c *commandService) SubmitAndWait(ctx context.Context, req *model.SubmitAndWaitRequest) (*model.SubmitAndWaitResponse, error) {
	protoReq := &v2.SubmitAndWaitRequest{
		Commands: commandsToProto(req.Commands),
	}

	resp, err := c.client.SubmitAndWait(ctx, protoReq)
	if err != nil {
		return nil, fmt.Errorf("failed to submit and wait: %w", err)
	}

	return &model.SubmitAndWaitResponse{
		UpdateID:         resp.UpdateId,
		CompletionOffset: resp.CompletionOffset,
	}, nil
}
