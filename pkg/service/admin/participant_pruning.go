package admin

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	adminv2 "github.com/digital-asset/dazl-client/v8/go/api/com/daml/ledger/api/v2/admin"
	"github.com/noders-team/go-daml/pkg/model"
)

type ParticipantPruning interface {
	Prune(ctx context.Context, pruneRequest *model.PruneRequest) error
}

type participantPruning struct {
	client adminv2.ParticipantPruningServiceClient
}

func NewParticipantPruningClient(conn *grpc.ClientConn) *participantPruning {
	client := adminv2.NewParticipantPruningServiceClient(conn)
	return &participantPruning{
		client: client,
	}
}

func (c *participantPruning) Prune(ctx context.Context, pruneRequest *model.PruneRequest) error {
	req := &adminv2.PruneRequest{
		PruneUpTo:                 pruneRequest.PruneUpTo,
		SubmissionId:              pruneRequest.SubmissionID,
		PruneAllDivulgedContracts: pruneRequest.PruneAllDivulgedContracts,
	}

	_, err := c.client.Prune(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to prune: %w", err)
	}

	return nil
}
