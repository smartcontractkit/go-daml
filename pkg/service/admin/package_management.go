package admin

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"

	adminv2 "github.com/digital-asset/dazl-client/v8/go/api/com/daml/ledger/api/v2/admin"
	"github.com/noders-team/go-daml/pkg/model"
)

type PackageManagement interface {
	ListKnownPackages(ctx context.Context) ([]*model.PackageDetails, error)
	UploadDarFile(ctx context.Context, darFile []byte, submissionID string) error
	ValidateDarFile(ctx context.Context, darFile []byte, submissionID string) error
}

type packageManagement struct {
	client adminv2.PackageManagementServiceClient
}

func NewPackageManagementClient(conn *grpc.ClientConn) *packageManagement {
	client := adminv2.NewPackageManagementServiceClient(conn)
	return &packageManagement{
		client: client,
	}
}

func (c *packageManagement) ListKnownPackages(ctx context.Context) ([]*model.PackageDetails, error) {
	req := &adminv2.ListKnownPackagesRequest{}

	resp, err := c.client.ListKnownPackages(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list known packages: %w", err)
	}

	return packageDetailsFromProtos(resp.PackageDetails), nil
}

func (c *packageManagement) UploadDarFile(ctx context.Context, darFile []byte, submissionID string) error {
	req := &adminv2.UploadDarFileRequest{
		DarFile:      darFile,
		SubmissionId: submissionID,
	}

	_, err := c.client.UploadDarFile(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to upload DAR file: %w", err)
	}

	return nil
}

func (c *packageManagement) ValidateDarFile(ctx context.Context, darFile []byte, submissionID string) error {
	req := &adminv2.ValidateDarFileRequest{
		DarFile:      darFile,
		SubmissionId: submissionID,
	}

	_, err := c.client.ValidateDarFile(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to validate DAR file: %w", err)
	}

	return nil
}

func packageDetailsFromProto(pb *adminv2.PackageDetails) *model.PackageDetails {
	if pb == nil {
		return nil
	}

	var knownSince *time.Time
	if pb.KnownSince != nil {
		t := pb.KnownSince.AsTime()
		knownSince = &t
	}

	return &model.PackageDetails{
		PackageID:   pb.PackageId,
		PackageSize: pb.PackageSize,
		KnownSince:  knownSince,
		Name:        pb.Name,
		Version:     pb.Version,
	}
}

func packageDetailsFromProtos(pbs []*adminv2.PackageDetails) []*model.PackageDetails {
	result := make([]*model.PackageDetails, len(pbs))
	for i, pb := range pbs {
		result[i] = packageDetailsFromProto(pb)
	}
	return result
}
