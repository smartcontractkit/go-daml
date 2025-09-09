package client

import (
	"github.com/noders-team/go-daml/pkg/service/admin"
	"github.com/noders-team/go-daml/pkg/service/ledger"
	"google.golang.org/grpc"
)

type damlBindingClient struct {
	client               *DamlClient
	grpcCl               *grpc.ClientConn
	UserMng              admin.UserManagement
	PartyMng             admin.PartyManagement
	PruningMng           admin.ParticipantPruning
	PackageMng           admin.PackageManagement
	CommandInspectionMng admin.CommandInspection
	IdentityProviderMng  admin.IdentityProviderConfig
	CommandCompletion    ledger.CommandCompletion
	CommandService       ledger.CommandService
	CommandSubmission    ledger.CommandSubmission
}

func NewDamlBindingClient(client *DamlClient, grpc *grpc.ClientConn) *damlBindingClient {
	return &damlBindingClient{
		client:               client,
		grpcCl:               grpc,
		UserMng:              admin.NewUserManagementClient(grpc),
		PartyMng:             admin.NewPartyManagementClient(grpc),
		PruningMng:           admin.NewParticipantPruningClient(grpc),
		PackageMng:           admin.NewPackageManagementClient(grpc),
		CommandInspectionMng: admin.NewCommandInspectionClient(grpc),
		IdentityProviderMng:  admin.NewIdentityProviderConfigClient(grpc),
		CommandCompletion:    ledger.NewCommandCompletionClient(grpc),
		CommandService:       ledger.NewCommandServiceClient(grpc),
		CommandSubmission:    ledger.NewCommandSubmissionClient(grpc),
	}
}

func (c *damlBindingClient) Close() {
	c.grpcCl.Close()
}
