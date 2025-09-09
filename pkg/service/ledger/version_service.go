package ledger

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	v2 "github.com/digital-asset/dazl-client/v8/go/api/com/daml/ledger/api/v2"
	"github.com/noders-team/go-daml/pkg/model"
)

type VersionService interface {
	GetLedgerAPIVersion(ctx context.Context, req *model.GetLedgerAPIVersionRequest) (*model.GetLedgerAPIVersionResponse, error)
}

type versionService struct {
	client v2.VersionServiceClient
}

func NewVersionServiceClient(conn *grpc.ClientConn) *versionService {
	client := v2.NewVersionServiceClient(conn)
	return &versionService{
		client: client,
	}
}

func (c *versionService) GetLedgerAPIVersion(ctx context.Context, req *model.GetLedgerAPIVersionRequest) (*model.GetLedgerAPIVersionResponse, error) {
	protoReq := &v2.GetLedgerApiVersionRequest{}

	resp, err := c.client.GetLedgerApiVersion(ctx, protoReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get ledger API version: %w", err)
	}

	return &model.GetLedgerAPIVersionResponse{
		Version:  resp.Version,
		Features: featuresFromProto(resp.Features),
	}, nil
}

func featuresFromProto(pb *v2.FeaturesDescriptor) *model.FeaturesDescriptor {
	if pb == nil {
		return nil
	}

	return &model.FeaturesDescriptor{
		UserManagement:   pb.UserManagement != nil && pb.UserManagement.Supported,
		PartyManagement:  pb.PartyManagement != nil,
		OffsetCheckpoint: pb.OffsetCheckpoint != nil,
	}
}
